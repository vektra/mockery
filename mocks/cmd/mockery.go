// Code generated by B2H-MockGen v0.0.0-dev. EDIT AT YOUR OWN PERIL.

package mocks

import (
	errors "github.com/pkg/errors"
	mock "github.com/stretchr/testify/mock"
)

type stackTracerMock struct {
	mock.Mock
}

func (_m *stackTracerMock) StackTrace() errors.StackTrace {
	args := _m.Called()
	return args.Get(0).(errors.StackTrace)
}