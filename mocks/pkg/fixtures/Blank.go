// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Blank is an autogenerated mock type for the Blank type
type Blank struct {
	mock.Mock
}

func (_m *Blank) On_Create(x interface{}, r_a0 error) *mock.Call {
	return _m.Mock.On("Create", x).Return(r_a0)
}

// Create provides a mock function with given fields: x
func (_m *Blank) Create(x interface{}) error {
	ret := _m.Called(x)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(x)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
