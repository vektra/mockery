// Code generated by mockery. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// GetInt is an autogenerated mock type for the GetInt type
type GetInt struct {
	mock.Mock
}

type GetInt_Expecter struct {
	mock *mock.Mock
}

func (_m *GetInt) EXPECT() *GetInt_Expecter {
	return &GetInt_Expecter{mock: &_m.Mock}
}

// Get provides a mock function with given fields:
func (_m *GetInt) Get() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// GetInt_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type GetInt_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
func (_e *GetInt_Expecter) Get() *GetInt_Get_Call {
	return &GetInt_Get_Call{Call: _e.mock.On("Get")}
}

func (_c *GetInt_Get_Call) Run(run func()) *GetInt_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *GetInt_Get_Call) Return(_a0 int) *GetInt_Get_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *GetInt_Get_Call) RunAndReturn(run func() int) *GetInt_Get_Call {
	_c.Call.Return(run)
	return _c
}

// NewGetInt creates a new instance of GetInt. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewGetInt(t interface {
	mock.TestingT
	Cleanup(func())
}) *GetInt {
	mock := &GetInt{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
