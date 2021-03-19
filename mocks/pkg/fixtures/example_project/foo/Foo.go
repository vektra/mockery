// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	foo "github.com/vektra/mockery/v2/pkg/fixtures/example_project/foo"
)

// Foo is an autogenerated mock type for the Foo type
type Foo struct {
	mock.Mock
}

func (_m *Foo) On_DoFoo(r_a0 string) *mock.Call {
	return _m.Mock.On("DoFoo").Return(r_a0)
}

// DoFoo provides a mock function with given fields:
func (_m *Foo) DoFoo() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

func (_m *Foo) On_GetBaz(r_a0 *foo.Baz, r_a1 error) *mock.Call {
	return _m.Mock.On("GetBaz").Return(r_a0, r_a1)
}

// GetBaz provides a mock function with given fields:
func (_m *Foo) GetBaz() (*foo.Baz, error) {
	ret := _m.Called()

	var r0 *foo.Baz
	if rf, ok := ret.Get(0).(func() *foo.Baz); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*foo.Baz)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
