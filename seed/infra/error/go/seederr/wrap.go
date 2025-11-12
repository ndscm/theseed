package seederr

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
)

type SeedError struct {
	code  uint32
	err   error
	stack []byte
}

func (err *SeedError) Code() uint32 {
	return err.code
}

func (err *SeedError) Error() string {
	if err.code != 0 {
		return fmt.Sprintf("[%d] %v\n\x1b[1;30m%v\x1b[0m", err.code, err.err, string(err.stack))
	}
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

func Code(code interface{}, err error) error {
	errCode := uint32(0)
	v := reflect.ValueOf(code)
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		errCode = uint32(v.Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		errCode = uint32(v.Int())
	default:
		errCode = 2 // UNKNOWN
	}
	seedErr := &SeedError{}
	if errors.As(err, &seedErr) {
		seedErr.code = errCode
		return seedErr
	}
	seedErr.code = errCode
	seedErr.err = err
	seedErr.stack = debug.Stack()
	return seedErr
}

func DefaultCode(code interface{}, err error) error {
	codeUint, ok := code.(uint32)
	if !ok {
		codeUint = 2 // UNKNOWN
	}
	seedErr := &SeedError{}
	if errors.As(err, &seedErr) {
		if seedErr.code == 0 {
			seedErr.code = codeUint
		}
		return seedErr
	}
	seedErr.code = codeUint
	seedErr.err = err
	seedErr.stack = debug.Stack()
	return seedErr
}

func WrapErrorf(format string, a ...any) error {
	return Wrap(fmt.Errorf(format, a...))
}

func CodeErrorf(code interface{}, format string, a ...any) error {
	return Code(code, fmt.Errorf(format, a...))
}
