// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
	mock "github.com/stretchr/testify/mock"
)

// NewRequesterSlice creates a new instance of RequesterSlice. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRequesterSlice(t interface {
	mock.TestingT
	Cleanup(func())
}) *RequesterSlice {
	mock := &RequesterSlice{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// RequesterSlice is an autogenerated mock type for the RequesterSlice type
type RequesterSlice struct {
	mock.Mock
}

type RequesterSlice_Expecter struct {
	mock *mock.Mock
}

func (_m *RequesterSlice) EXPECT() *RequesterSlice_Expecter {
	return &RequesterSlice_Expecter{mock: &_m.Mock}
}

// Get provides a mock function for the type RequesterSlice
func (_mock *RequesterSlice) Get(pathParam string) ([]string, error) {
	ret := _mock.Called(pathParam)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 []string
	var r1 error
	if returnFunc, ok := ret.Get(0).(func(string) ([]string, error)); ok {
		return returnFunc(pathParam)
	}
	if returnFunc, ok := ret.Get(0).(func(string) []string); ok {
		r0 = returnFunc(pathParam)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(string) error); ok {
		r1 = returnFunc(pathParam)
	} else {
		r1 = ret.Error(1)
	}
	return r0, r1
}

// RequesterSlice_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type RequesterSlice_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - pathParam
func (_e *RequesterSlice_Expecter) Get(pathParam interface{}) *RequesterSlice_Get_Call {
	return &RequesterSlice_Get_Call{Call: _e.mock.On("Get", pathParam)}
}

func (_c *RequesterSlice_Get_Call) Run(run func(pathParam string)) *RequesterSlice_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *RequesterSlice_Get_Call) Return(stringsOutParam []string, errOutParam error) *RequesterSlice_Get_Call {
	_c.Call.Return(stringsOutParam, errOutParam)
	return _c
}

func (_c *RequesterSlice_Get_Call) RunAndReturn(run func(pathParam string) ([]string, error)) *RequesterSlice_Get_Call {
	_c.Call.Return(run)
	return _c
}
