// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Requester4 is an autogenerated mock type for the Requester4 type
type Requester4 struct {
	mock.Mock
}

func (_m *Requester4) On_Get() *mock.Call {
	return _m.Mock.On("Get").Return()
}

// Get provides a mock function with given fields:
func (_m *Requester4) Get() {
	_m.Called()
}
