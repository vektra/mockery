// Code generated by B2H-MockGen v0.0.0-dev. EDIT AT YOUR OWN PERIL.

package b2hmock

import mock "github.com/stretchr/testify/mock"

type CleanupMock struct {
	mock.Mock
}

func (_m *CleanupMock) Execute() error {
	args := _m.Called()
	return args.Error(0)
}