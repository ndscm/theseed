package seederr

import (
	"errors"
	"fmt"
	"runtime/debug"
)

type SeedError struct {
	err   error
	stack []byte
}

func (err *SeedError) Error() string {
	return fmt.Sprintf("%v\n\x1b[1;30m%v\x1b[0m", err.err, string(err.stack))
}

func (err *SeedError) Unwrap() error {
	return err.err
}

func Wrap(err error) error {
	seedErr := &SeedError{}
	if errors.As(err, &seedErr) {
		return seedErr
	}
	seedErr.err = err
	seedErr.stack = debug.Stack()
	return seedErr
}

func WrapErrorf(format string, a ...any) error {
	return Wrap(fmt.Errorf(format, a...))
}
