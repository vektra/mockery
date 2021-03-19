// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	test "github.com/vektra/mockery/v2/pkg/fixtures"
)

// A is an autogenerated mock type for the A type
type A struct {
	mock.Mock
}

func (_m *A) On_Call(r_a0 test.B, r_a1 error) *mock.Call {
	return _m.Mock.On("Call").Return(r_a0, r_a1)
}

// Call provides a mock function with given fields:
func (_m *A) Call() (test.B, error) {
	ret := _m.Called()

	var r0 test.B
	if rf, ok := ret.Get(0).(func() test.B); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(test.B)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
