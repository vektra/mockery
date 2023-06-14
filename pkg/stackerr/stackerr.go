package stackerr

import (
	"errors"
	"fmt"
	"runtime/debug"
)

type StackErr struct {
	cause error
	stack []byte
}

func NewStackErr(cause error) error {
	return StackErr{
		cause: cause,
		stack: debug.Stack(),
	}
}

func NewStackErrf(cause error, f string, args ...any) error {
	msg := fmt.Sprintf(f, args...)
	cause = fmt.Errorf(msg+": %w", cause)
	return NewStackErr(cause)
}

func (se StackErr) Error() string {
	return se.cause.Error()
}

func (se StackErr) Unwrap() error {
	return se.cause
}

func (se StackErr) Stack() []byte {
	return se.stack
}

func GetStack(err error) ([]byte, bool) {
	var s interface {
		Stack() []byte
	}
	if errors.As(err, &s) {
		return s.Stack(), true
	}
	return nil, false
}
