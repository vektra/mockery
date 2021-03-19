// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Requester is an autogenerated mock type for the Requester type
type Requester struct {
	mock.Mock
}

func (_m *Requester) On_Get(path string, r_a0 string, r_a1 error) *mock.Call {
	return _m.Mock.On("Get", path).Return(r_a0, r_a1)
}

// Get provides a mock function with given fields: path
func (_m *Requester) Get(path string) (string, error) {
	ret := _m.Called(path)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
