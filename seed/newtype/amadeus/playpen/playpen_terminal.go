package playpen

import (
	"os"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// defaultTtyRows and defaultTtyCols size a pseudo-terminal whose caller did not
// ask for a size. A 0x0 terminal is legal but useless: the shell cannot place
// its prompt and full-screen programs draw nothing, so fall back to the
// conventional dumb-terminal geometry until the first Resize.
const defaultTtyRows = 24
const defaultTtyCols = 80

// PlaypenTerminal is a PlaypenShell whose shell is attached to a
// pseudo-terminal, as if someone had sat down in front of it.
//
// The difference from the plain session it embeds is what the shell believes.
// A plain session runs `podman exec --interactive` and the shell sees no
// terminal: zsh -i starts with no prompt, no echo, and no job control, which is
// what makes it good for driving a program. A terminal session adds `--tty`, so
// the shell sources ~/.zshrc and then behaves the way it does for a human — it
// prints a prompt, echoes keystrokes, interprets control characters, and merges
// stdout and stderr onto the one terminal.
//
// That merge is why Stderr is nil here, and why Stdin and Stdout are the same
// file: a terminal is one stream, read and written at the same end. Stdout
// carries raw terminal output, escape sequences included, so a caller must feed
// it to a terminal emulator rather than treat it as text; writing to Stdin is
// indistinguishable, to the shell, from typing.
//
// Delegate is inherited but has no business here: it exec-replaces the shell by
// writing a launch line to its stdin, which on a terminal would be echoed back
// and would leave the caller talking to the program instead of the shell.
type PlaypenTerminal struct {
	PlaypenShell

	// ptyFile is the master side of the pseudo-terminal, and is what both
	// Stdin and Stdout point at.
	ptyFile *os.File
}

// Resize changes the terminal's window size, which delivers SIGWINCH to the
// foreground process group inside the container so the shell and any
// full-screen program it is running redraw at the new size.
//
// The size travels the whole way: podman is attached to this pseudo-terminal,
// so it observes the resize on its own controlling terminal and applies it to
// the terminal it allocated in the container.
func (pt *PlaypenTerminal) Resize(rows uint16, cols uint16) error {
	if rows == 0 || cols == 0 {
		return seederr.WrapErrorf("playpen tty: refusing zero window size %dx%d", rows, cols)
	}
	err := pty.Setsize(pt.ptyFile, &pty.Winsize{Rows: rows, Cols: cols})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// Close ends the session, overriding PlaypenShell.Close: stdin EOF is
// meaningless to a shell reading from a terminal.
//
// Hanging up the terminal is not enough on its own either. The process on the
// slave side of this pseudo-terminal is the podman exec client, not the shell —
// the shell sits inside the container on a terminal podman allocated for it —
// so closing the master gives podman an I/O error on a stream it is not obliged
// to care about, and it keeps running. Close therefore hangs up and then
// signals the exec client, which is what actually ends the session, falling
// back to a kill if the signal is ignored.
//
// As with the plain session, ending the host-side client does not reliably reap
// the process inside the container on every podman version;
// PlaypenController.Shutdown (podman rm --force) remains the backstop.
func (pt *PlaypenTerminal) Close() error {
	// Long enough for podman to tear down its side of the exec, short enough
	// that a caller closing a terminal does not sit and wait on it.
	const closeGracePeriod = 5 * time.Second

	err := pt.ptyFile.Close()
	if err != nil {
		seedlog.Warnf("playpen tty: error closing pseudo-terminal: %v", err)
	}

	termErr := pt.cmd.Process.Signal(syscall.SIGTERM)
	if termErr != nil {
		seedlog.Warnf("playpen tty: error terminating exec client: %v", termErr)
	}

	waited := make(chan struct{})
	go func() {
		defer close(waited)
		pt.Wait()
	}()

	select {
	case <-waited:
		// The shell's exit status is not interesting here: a terminal whose
		// window is closed reports whatever the shell happened to be doing,
		// and the caller asked for it to end.
	case <-time.After(closeGracePeriod):
		killErr := pt.cmd.Process.Kill()
		if killErr != nil {
			seedlog.Warnf("playpen tty: error killing exec client: %v", killErr)
		}
		<-waited
		return seederr.WrapErrorf(
			"playpen tty: exec client did not exit within %s of SIGTERM; killed it",
			closeGracePeriod)
	}
	return nil
}
