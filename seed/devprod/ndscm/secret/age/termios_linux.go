//go:build linux

package age

import (
	"golang.org/x/sys/unix"
)

// Linux uses the TCGETS/TCSETS ioctls to get and set terminal attributes.
const (
	ioctlGetTermios = unix.TCGETS
	ioctlSetTermios = unix.TCSETS
)
