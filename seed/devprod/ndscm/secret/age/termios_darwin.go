//go:build darwin

package age

import (
	"golang.org/x/sys/unix"
)

// Darwin (like the BSDs) uses the TIOCGETA/TIOCSETA ioctls for terminal attributes.
const (
	ioctlGetTermios = unix.TIOCGETA
	ioctlSetTermios = unix.TIOCSETA
)
