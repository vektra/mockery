// Code generated by mockery. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// RequesterPtr is an autogenerated mock type for the RequesterPtr type
type RequesterPtr struct {
	mock.Mock
}

type RequesterPtr_Expecter struct {
	mock *mock.Mock
}

func (_m *RequesterPtr) EXPECT() *RequesterPtr_Expecter {
	return &RequesterPtr_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields: path
func (_m *RequesterPtr) Get(path string) (*string, error) {
	ret := _m.Called(path)

	var r0 *string
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*string, error)); ok {
		return rf(path)
	}
	if rf, ok := ret.Get(0).(func(string) *string); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RequesterPtr_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type RequesterPtr_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - path string
func (_e *RequesterPtr_Expecter) Get(path interface{}) *RequesterPtr_Get_Call {
	return &RequesterPtr_Get_Call{Call: _e.mock.On("Get", path)}
}

func (_c *RequesterPtr_Get_Call) Run(run func(path string)) *RequesterPtr_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *RequesterPtr_Get_Call) Return(_a0 *string, _a1 error) *RequesterPtr_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *RequesterPtr_Get_Call) RunAndReturn(run func(string) (*string, error)) *RequesterPtr_Get_Call {
	_c.Call.Return(run)
	return _c
}

// NewRequesterPtr creates a new instance of RequesterPtr. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRequesterPtr(t interface {
	mock.TestingT
	Cleanup(func())
}, expectedCalls ...*mock.Call) *RequesterPtr {
	mock := &RequesterPtr{}
	mock.Mock.Test(t)
	mock.Mock.ExpectedCalls = expectedCalls

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
