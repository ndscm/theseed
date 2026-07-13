// Package playpen manages the emulated workstation container in which a human
// or agent does its work. A PlaypenController owns a single podman container
// running systemd as init; tty sessions opened against it run a program as the
// playpen user, sharing the container's filesystem and process namespace.
package playpen

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// playpenContainerName is the fixed name of the playpen container. Only one runs
// per controller, so a stable name lets exec and teardown target it without
// threading an id around.
const playpenContainerName = "playpen"

// PlaypenController owns the running playpen container and hands out shell
// sessions into it.
type PlaypenController struct {
	userHandle string
}

// Home is the playpen user's home directory inside the container. The shared
// playpen home is mounted at /home, so the login user's home is /home/<handle>.
func (pc *PlaypenController) Home() string {
	return "/home/" + pc.userHandle
}

// StartShell opens a shell session inside the playpen container, running shellPath
// with args as the playpen user (e.g. "/usr/bin/zsh", ["-i"]). It returns a
// PlaypenShell whose Stdin, Stdout, and Stderr are wired to the shell; call
// Delegate to hand the session to a long-lived program.
//
// The session is started under ctx: cancelling ctx kills the podman exec client
// on the host. Whether that also terminates the program inside the container is
// podman/conmon-version dependent and not guaranteed — on some versions the
// exec'd process is orphaned and keeps running under the container's systemd — so
// ctx cancellation is a backstop, not a reliable reaper. Prefer a graceful
// shutdown: Close (or closing Stdin) delivers stdin EOF, which a program that
// quits on EOF (e.g. claude --print) treats as its cue to finish and exit, and
// the client then observes that clean exit and reports it through Wait. EOF is
// only a cue, not a kill: a program that ignores it will not exit on its own, so
// Close bounds the wait and then kills. The container's own teardown
// (PlaypenController.Shutdown, i.e. podman rm --force) is the final backstop that
// removes anything left behind.
func (pc *PlaypenController) StartShell(
	ctx context.Context, shellPath string, args []string,
) (*PlaypenShell, error) {
	execArgs := []string{
		"exec",
		"--interactive",
		"--user", pc.userHandle,
		playpenContainerName,
		shellPath,
	}
	execArgs = append(execArgs, args...)

	cmd := exec.CommandContext(ctx, "podman", execArgs...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	err = cmd.Start()
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	seedlog.Infof("Playpen shell session started. pid=%d shell=%s %v user=%s",
		cmd.Process.Pid, shellPath, args, pc.userHandle)

	return &PlaypenShell{
		containerName: playpenContainerName,

		userHandle: pc.userHandle,

		cmd: cmd,

		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}, nil
}

// StartTerminal opens a terminal session inside the playpen container,
// running shellPath with args as the playpen user (e.g. "/usr/bin/zsh", ["-i"])
// on a pseudo-terminal sized to window. A zero row or column count falls back
// to the conventional 24x80; window's pixel dimensions are passed through
// untouched, and zero is the right value for them on a terminal nobody is
// drawing graphics on.
//
// The size is applied to the terminal before the exec client starts, not after:
// podman reads the size of the terminal it is attached to in order to size the
// terminal it allocates in the container. Starting at 0x0 and resizing
// afterwards would race that read, and a lost resize would leave the shell
// convinced it has no window at all.
//
// The session is started under ctx, with the same caveat as StartShellSession:
// cancelling ctx kills the podman exec client on the host, but whether that
// terminates the shell inside the container is podman-version dependent. Prefer
// Close, which hangs up the terminal.
func (pc *PlaypenController) StartTerminal(
	ctx context.Context, window pty.Winsize, shellPath string, args []string,
) (*PlaypenTerminal, error) {
	if window.Rows == 0 {
		window.Rows = defaultTtyRows
	}
	if window.Cols == 0 {
		window.Cols = defaultTtyCols
	}

	// --tty asks podman to allocate a pseudo-terminal for the shell inside the
	// container, which it sizes from the terminal its own client is attached to.
	execArgs := []string{
		"exec",
		"--interactive",
		"--tty",
		"--user", pc.userHandle,
		playpenContainerName,
		shellPath,
	}
	execArgs = append(execArgs, args...)

	cmd := exec.CommandContext(ctx, "podman", execArgs...)

	// StartWithSize hands podman a pseudo-terminal as its stdin, stdout, and
	// stderr, already sized, and starts it.
	ptyFile, err := pty.StartWithSize(cmd, &window)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	seedlog.Infof("Playpen tty session started. pid=%d shell=%s %v user=%s size=%dx%d",
		cmd.Process.Pid, shellPath, args, pc.userHandle, window.Rows, window.Cols)

	session := &PlaypenTerminal{
		ptyFile: ptyFile,
	}
	session.userHandle = pc.userHandle
	session.cmd = cmd
	session.Stdin = ptyFile
	session.Stdout = ptyFile
	return session, nil
}

func (pc *PlaypenController) SimpleFileSystem() *SimpleFileSystem {
	return WrapSimpleFileSystem(playpenContainerName, pc.userHandle)
}

// Shutdown stops and removes the playpen container. It is best effort: --force
// sends the container's stop signal, waits, then removes it.
func (pc *PlaypenController) Shutdown() error {
	cmd := exec.Command("podman", "rm", "--force", playpenContainerName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// BootPlaypenController starts the playpen container detached and returns a
// controller for it. The container runs systemd as init under a private cgroup
// namespace parented at /amadeus, with the shared playpen home mounted in, and
// creates userHandle as its login user.
func BootPlaypenController(userHandle string) (*PlaypenController, error) {
	const playpenImage = "ghcr.io/ndscm/seed-newtype-amadeus-playpen-container:latest"

	cmd := exec.Command("podman", "run",
		"--name="+playpenContainerName,
		"--replace",
		"--detach",
		"--network=host",
		"--pull=never",
		"--systemd=always",
		"--cgroupns=private",
		"--cgroups=enabled",
		"--cgroup-parent=/amadeus",
		"-v", "/playpen/home:/home:Z",
		playpenImage,
		userHandle,
		"systemd",
	)
	runOutput, err := cmd.Output()
	if err != nil {
		stderr := ""
		exitErr, ok := errors.AsType[*exec.ExitError](err)
		if ok {
			stderr = string(exitErr.Stderr)
		}
		return nil, seederr.WrapErrorf("podman run playpen failed: %v: %s", err, stderr)
	}
	seedlog.Infof("Playpen container started. output=%s", runOutput)

	// podman run --detach reports the container up the instant its entrypoint
	// starts, but the entrypoint runs useradd before it execs systemd; until the
	// user lands in passwd a StartShell's `podman exec --user` fails with
	// "unable to find user in passwd". Poll with the same `podman exec --user`
	// StartShell uses so sessions are only handed out once the user is ready.
	const readyTimeout = 30 * time.Second
	const readyInterval = 200 * time.Millisecond
	deadline := time.Now().Add(readyTimeout)
	for {
		probe := exec.Command("podman", "exec", "--user", userHandle,
			playpenContainerName, "id", userHandle)
		_, err := probe.Output()
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			stderr := ""
			exitErr, ok := errors.AsType[*exec.ExitError](err)
			if ok {
				stderr = string(exitErr.Stderr)
			}
			return nil, seederr.WrapErrorf(
				"playpen user %q not ready after %s: %v: %s",
				userHandle, readyTimeout, err, stderr)
		}
		time.Sleep(readyInterval)
	}

	seedlog.Infof("Playpen container booted. user=%s", userHandle)

	return &PlaypenController{
		userHandle: userHandle,
	}, nil
}
