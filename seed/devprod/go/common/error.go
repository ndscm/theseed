package common

import (
	"fmt"
	"runtime/debug"
)

func WrapTrace(err error) error {
	return fmt.Errorf("%w\n\x1b[1;30m%v\x1b[0m", err, string(debug.Stack()))
}
