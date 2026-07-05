// Package playpen manages the emulated workstation container in which a human
// or agent does its work. A PlaypenController owns a single podman container
// running systemd as init; tty sessions opened against it run a program as the
// playpen user, sharing the container's filesystem and process namespace.
package playpen

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// playpenContainerName is the fixed name of the playpen container. Only one runs
// per controller, so a stable name lets exec and teardown target it without
// threading an id around.
const playpenContainerName = "playpen"

// shellQuote wraps s in single quotes so the shell treats it as one literal
// word, escaping any embedded single quotes. It is used to build the launch
// line Delegate feeds to the shell so a workdir or argument can't be split or
// interpreted.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// PlaypenTty is a live shell session inside the playpen container: a persistent
// process, started as the playpen user's login shell (so ~/.zshrc is sourced and
// its PATH and environment are in effect). Delegate then hands the session's
// Stdin, Stdout, and Stderr to a single long-lived program by exec-replacing the
// shell with it. The program owns the streams for its whole life, so a caller can
// push input to it continuously (e.g. a claude subprocess in stream-json mode)
// and read its output as it is produced.
//
// The session runs under `podman exec --interactive` without a pseudo-terminal.
// That keeps stdout and stderr as separate streams (a PTY would merge them), at
// the cost of the shell not seeing a real terminal: zsh -i starts without job
// control or a prompt, which is what we want for programmatic driving.
type PlaypenTty struct {
	// userHandle is the playpen user the session runs as, retained so Delegate
	// can prepare the workdir out-of-band as that same user.
	userHandle string

	cmd *exec.Cmd

	// Stdin, Stdout, and Stderr are the shell's own streams; after Delegate the
	// caller drives the delegated program through them directly.
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser
}

// Pid reports the pid of the podman exec client on the host. The program inside
// the container runs as (and, after Delegate's exec, is) the child of that
// client.
func (pt *PlaypenTty) Pid() int {
	return pt.cmd.Process.Pid
}

// Wait blocks until the session's process exits and returns its exit error, if
// any. Use Close instead to actively end the session by closing its stdin.
func (pt *PlaypenTty) Wait() error {
	err := pt.cmd.Wait()
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// Close ends the session by closing the shell's stdin, which delivers EOF so a
// program that quits on EOF (the shell, or a delegated program like claude
// --print) exits, then reaps it. EOF is only a cue: a program that ignores it
// would leave Close blocked in Wait forever, so Close bounds the wait by
// closeGracePeriod and then kills the host exec client to guarantee it returns.
// Killing the client may not reap the in-container process on every podman
// version — Shutdown (podman rm --force) is the final backstop for that.
//
// Callers that need an in-flight turn to finish should drain first (e.g. wait
// for the program to go idle) before calling Close; once called, Close will
// force the session down after the grace period.
func (pt *PlaypenTty) Close() error {
	const closeGracePeriod = 30 * time.Second

	err := pt.Stdin.Close()
	if err != nil {
		seedlog.Warnf("playpen tty: error closing stdin: %v", err)
	}

	waited := make(chan error, 1)
	go func() {
		waited <- pt.cmd.Wait()
	}()

	select {
	case waitErr := <-waited:
		if waitErr != nil {
			return seederr.Wrap(waitErr)
		}
	case <-time.After(closeGracePeriod):
		killErr := pt.cmd.Process.Kill()
		if killErr != nil {
			seedlog.Warnf("playpen tty: error killing exec client: %v", killErr)
		}
		<-waited
		return seederr.WrapErrorf(
			"playpen tty: program did not exit on stdin EOF within %s; killed exec client",
			closeGracePeriod)
	}
	return nil
}

// Delegate exec-replaces the session's shell with command (with args), run in
// workdir, handing it the session's Stdin, Stdout, and Stderr for its whole
// life. workdir is created first — as the playpen user, inside the container —
// so callers need not pre-create it. After Delegate the caller drives the
// program through Stdin, Stdout, and Stderr and reaps it with Wait. Delegate
// writes the launch line and returns; it does not wait for the program to
// finish.
func (pt *PlaypenTty) Delegate(workdir string, command string, args []string) error {
	// Create the workdir out-of-band, as the playpen user inside the container,
	// so an uncreatable workdir surfaces as a real error here. The program is
	// launched by writing an exec line to the shell's stdin, which reports
	// nothing back: if the inline `cd` below failed instead, the exec would
	// never run, the bare shell would keep reading stdin, and every subsequent
	// write would be fed to it rather than the program — with Wait blocking
	// forever and no error surfaced. Verifying creatability up front turns that
	// silent failure into a synchronous one.
	mkdir := exec.Command("podman", "exec", "--user", pt.userHandle,
		playpenContainerName, "mkdir", "-p", workdir)
	_, err := mkdir.Output()
	if err != nil {
		stderr := ""
		exitErr, ok := errors.AsType[*exec.ExitError](err)
		if ok {
			stderr = string(exitErr.Stderr)
		}
		return seederr.WrapErrorf(
			"create playpen workdir %q failed: %v: %s", workdir, err, stderr)
	}

	launch := []string{"exec", shellQuote(command)}
	for _, arg := range args {
		launch = append(launch, shellQuote(arg))
	}

	// cd guards the exec: the program only starts once its working directory is
	// entered. exec replaces the shell, so no marker or prompt follows the
	// program's own output on stdout.
	quotedWorkdir := shellQuote(workdir)
	script := fmt.Sprintf("cd %s && %s\n", quotedWorkdir, strings.Join(launch, " "))

	_, err = io.WriteString(pt.Stdin, script)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

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

// StartTty opens a shell session inside the playpen container, running shellPath
// with args as the playpen user (e.g. "/usr/bin/zsh", ["-i"]). It returns a
// PlaypenTty whose Stdin, Stdout, and Stderr are wired to the shell; call
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
func (pc *PlaypenController) StartTty(
	ctx context.Context, shellPath string, args []string,
) (*PlaypenTty, error) {
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

	seedlog.Infof("Playpen tty started. pid=%d shell=%s %v user=%s",
		cmd.Process.Pid, shellPath, args, pc.userHandle)

	return &PlaypenTty{
		userHandle: pc.userHandle,

		cmd: cmd,

		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}, nil
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
	// user lands in passwd a StartTty's `podman exec --user` fails with "unable
	// to find user in passwd". Poll with the same `podman exec --user` StartTty
	// uses so sessions are only handed out once the user is ready.
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
