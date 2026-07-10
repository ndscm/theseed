package playpen

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// shellQuote wraps s in single quotes so the shell treats it as one literal
// word, escaping any embedded single quotes. It is used to build the launch
// line Delegate feeds to the shell so a workdir or argument can't be split or
// interpreted.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// PlaypenShell is a live shell session inside the playpen container: a persistent
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
// control or a prompt, which is what we want for programmatic driving. A caller
// that wants the shell to behave as it does for a human — prompt, echo, job
// control — wants PlaypenTerminal instead.
type PlaypenShell struct {
	containerName string

	// userHandle is the playpen user the session runs as, retained so Delegate
	// can prepare the workdir out-of-band as that same user.
	userHandle string

	cmd *exec.Cmd

	// Stdin, Stdout, and Stderr are the shell's own streams; after Delegate the
	// caller drives the delegated program through them directly.
	Stdin  io.WriteCloser
	Stdout io.ReadCloser
	Stderr io.ReadCloser

	// waitOnce makes Wait idempotent, and so safe to call from more than one
	// goroutine: exec.Cmd.Wait may only run once. PlaypenTerminal needs that,
	// because the goroutine draining its terminal reaps the shell and so does
	// Close.
	waitOnce sync.Once
	waitErr  error
}

// Pid reports the pid of the podman exec client on the host. The program inside
// the container runs as (and, after Delegate's exec, is) the child of that
// client.
func (pt *PlaypenShell) Pid() int {
	return pt.cmd.Process.Pid
}

// Wait blocks until the session's process exits and returns its exit error, if
// any. Use Close instead to actively end the session. It is safe to call from
// several goroutines and more than once; every caller sees the same status.
func (pt *PlaypenShell) Wait() error {
	pt.waitOnce.Do(func() {
		pt.waitErr = pt.cmd.Wait()
	})
	if pt.waitErr != nil {
		return seederr.Wrap(pt.waitErr)
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
//
// PlaypenTerminal overrides this: stdin EOF means nothing to a terminal.
func (pt *PlaypenShell) Close() error {
	const closeGracePeriod = 30 * time.Second

	err := pt.Stdin.Close()
	if err != nil {
		seedlog.Warnf("playpen shell session: error closing stdin: %v", err)
	}

	waited := make(chan error, 1)
	go func() {
		waited <- pt.Wait()
	}()

	select {
	case waitErr := <-waited:
		if waitErr != nil {
			return seederr.Wrap(waitErr)
		}
	case <-time.After(closeGracePeriod):
		killErr := pt.cmd.Process.Kill()
		if killErr != nil {
			seedlog.Warnf("playpen shell session: error killing exec client: %v", killErr)
		}
		<-waited
		return seederr.WrapErrorf(
			"playpen shell session: program did not exit on stdin EOF within %s; killed exec client",
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
func (pt *PlaypenShell) Delegate(workdir string, command string, args []string) error {
	// Create the workdir out-of-band, as the playpen user inside the container,
	// so an uncreatable workdir surfaces as a real error here. The program is
	// launched by writing an exec line to the shell's stdin, which reports
	// nothing back: if the inline `cd` below failed instead, the exec would
	// never run, the bare shell would keep reading stdin, and every subsequent
	// write would be fed to it rather than the program — with Wait blocking
	// forever and no error surfaced. Verifying creatability up front turns that
	// silent failure into a synchronous one.
	mkdir := exec.Command("podman", "exec", "--user", pt.userHandle,
		pt.containerName, "mkdir", "-p", workdir)
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
