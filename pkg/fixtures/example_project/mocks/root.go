// Code generated by B2H-MockGen v0.0.0-dev. EDIT AT YOUR OWN PERIL.

package mocks

import (
	foo "github.com/pendo-io/b2h-mockgen/pkg/fixtures/example_project/foo"
	mock "github.com/stretchr/testify/mock"
)

type RootMock struct {
	mock.Mock
}

func (_m *RootMock) ReturnsFoo() (foo.Foo, error) {
	args := _m.Called()
	return args.Get(0).(foo.Foo), args.Error(1)
}
func (_m *RootMock) TakesBaz(_a0 *foo.Baz) {
	_m.Called(_a0)
}
