// Code generated by B2H-MockGen v0.0.0-dev. EDIT AT YOUR OWN PERIL.

package b2hmock

import mock "github.com/stretchr/testify/mock"

type RequesterElidedMock struct {
	mock.Mock
}

func (_m *RequesterElidedMock) Get(path string, url string) error {
	args := _m.Called(path, url)
	return args.Error(0)
}