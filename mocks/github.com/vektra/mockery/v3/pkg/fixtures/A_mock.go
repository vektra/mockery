
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
	"github.com/vektra/mockery/v3/pkg/fixtures"
    mock "github.com/stretchr/testify/mock"
)

 
// NewA creates a new instance of A. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewA (t interface {
	mock.TestingT
	Cleanup(func())
}) *A {
	mock := &A{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}


// A is an autogenerated mock type for the A type
type A struct {
	mock.Mock
}

type A_Expecter struct {
	mock *mock.Mock
}

func (_m *A) EXPECT() *A_Expecter {
	return &A_Expecter{mock: &_m.Mock}
}

 

// Call provides a mock function for the type A
func (_mock *A) Call() (test.B, error) {  
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Call")
	}

		
	var r0 test.B
	var r1 error
	if returnFunc, ok := ret.Get(0).(func() (test.B, error)); ok {
		return returnFunc()
	} 
	if returnFunc, ok := ret.Get(0).(func() test.B); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(test.B)
	} 
	if returnFunc, ok := ret.Get(1).(func() error); ok {
		r1 = returnFunc()
	} else {
		r1 = ret.Error(1)
	} 
	return r0, r1
}



// A_Call_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Call'
type A_Call_Call struct {
	*mock.Call
}



// Call is a helper method to define mock.On call
func (_e *A_Expecter) Call() *A_Call_Call {
	return &A_Call_Call{Call: _e.mock.On("Call", )}
}

func (_c *A_Call_Call) Run(run func()) *A_Call_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *A_Call_Call) Return(b test.B, err error) *A_Call_Call {
	_c.Call.Return(b, err)
	return _c
}

func (_c *A_Call_Call) RunAndReturn(run func()(test.B, error)) *A_Call_Call {
	_c.Call.Return(run)
	return _c
}
  
