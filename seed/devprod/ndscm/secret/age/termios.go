package age

import (
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/sys/unix"
)

// disableEcho turns off terminal echo on fd, returning a function that restores
// the previous terminal state. Canonical line editing is preserved, so the
// secret can still be typed and edited; it simply does not appear on screen.
func disableEcho(fd uintptr) (func(), error) {
	termios, err := unix.IoctlGetTermios(int(fd), ioctlGetTermios)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	restore := func() {
		err := unix.IoctlSetTermios(int(fd), ioctlSetTermios, termios)
		if err != nil {
			seedlog.Warnf("failed to restore terminal echo: %v", err)
		}
	}
	updated := *termios
	updated.Lflag &^= unix.ECHO
	err = unix.IoctlSetTermios(int(fd), ioctlSetTermios, &updated)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return restore, nil
}
