package stackerr

import (
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
