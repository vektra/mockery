// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
	mock "github.com/stretchr/testify/mock"
)

// NewRequesterArgSameAsPkg creates a new instance of RequesterArgSameAsPkg. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRequesterArgSameAsPkg(t interface {
	mock.TestingT
	Cleanup(func())
}) *RequesterArgSameAsPkg {
	mock := &RequesterArgSameAsPkg{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// RequesterArgSameAsPkg is an autogenerated mock type for the RequesterArgSameAsPkg type
type RequesterArgSameAsPkg struct {
	mock.Mock
}

type RequesterArgSameAsPkg_Expecter struct {
	mock *mock.Mock
}

func (_m *RequesterArgSameAsPkg) EXPECT() *RequesterArgSameAsPkg_Expecter {
	return &RequesterArgSameAsPkg_Expecter{mock: &_m.Mock}
}

// Get provides a mock function for the type RequesterArgSameAsPkg
func (_mock *RequesterArgSameAsPkg) Get(testParam string) {
	_mock.Called(testParam)
	return
}

// RequesterArgSameAsPkg_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type RequesterArgSameAsPkg_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - testParam
func (_e *RequesterArgSameAsPkg_Expecter) Get(testParam interface{}) *RequesterArgSameAsPkg_Get_Call {
	return &RequesterArgSameAsPkg_Get_Call{Call: _e.mock.On("Get", testParam)}
}

func (_c *RequesterArgSameAsPkg_Get_Call) Run(run func(testParam string)) *RequesterArgSameAsPkg_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *RequesterArgSameAsPkg_Get_Call) Return() *RequesterArgSameAsPkg_Get_Call {
	_c.Call.Return()
	return _c
}

func (_c *RequesterArgSameAsPkg_Get_Call) RunAndReturn(run func(testParam string)) *RequesterArgSameAsPkg_Get_Call {
	_c.Run(run)
	return _c
}
