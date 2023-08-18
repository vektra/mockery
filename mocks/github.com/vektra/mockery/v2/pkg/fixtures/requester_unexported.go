// Code generated by mockery. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// requester_unexported is an autogenerated mock type for the requester_unexported type
type requester_unexported struct {
	mock.Mock
}

type requester_unexported_Expecter struct {
	mock *mock.Mock
}

func (_m *requester_unexported) EXPECT() *requester_unexported_Expecter {
	return &requester_unexported_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields:
func (_m *requester_unexported) Get() {
	_m.Called()
}

// requester_unexported_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type requester_unexported_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
func (_e *requester_unexported_Expecter) Get() *requester_unexported_Get_Call {
	return &requester_unexported_Get_Call{Call: _e.mock.On("Get")}
}

func (_c *requester_unexported_Get_Call) Run(run func()) *requester_unexported_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *requester_unexported_Get_Call) Return() *requester_unexported_Get_Call {
	_c.Call.Return()
	return _c
}

func (_c *requester_unexported_Get_Call) RunAndReturn(run func()) *requester_unexported_Get_Call {
	_c.Call.Return(run)
	return _c
}

// newRequester_unexported creates a new instance of requester_unexported. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newRequester_unexported(t interface {
	mock.TestingT
	Cleanup(func())
}, expectedCalls ...*mock.Call) *requester_unexported {
	mock := &requester_unexported{}
	mock.Mock.Test(t)
	mock.Mock.ExpectedCalls = expectedCalls

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
