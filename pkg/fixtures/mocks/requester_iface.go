// Code generated by B2H-MockGen v0.0.0-dev. EDIT AT YOUR OWN PERIL.

package mocks

import (
	io "io"

	mock "github.com/stretchr/testify/mock"
)

type RequesterIfaceMock struct {
	mock.Mock
}

func (_m *RequesterIfaceMock) Get() io.Reader {
	args := _m.Called()
	return args.Get(0).(io.Reader)
}
