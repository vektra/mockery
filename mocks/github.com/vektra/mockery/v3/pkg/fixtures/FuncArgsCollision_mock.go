
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
    mock "github.com/stretchr/testify/mock"
)

 
// NewFuncArgsCollision creates a new instance of FuncArgsCollision. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFuncArgsCollision (t interface {
	mock.TestingT
	Cleanup(func())
}) *FuncArgsCollision {
	mock := &FuncArgsCollision{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}


// FuncArgsCollision is an autogenerated mock type for the FuncArgsCollision type
type FuncArgsCollision struct {
	mock.Mock
}

type FuncArgsCollision_Expecter struct {
	mock *mock.Mock
}

func (_m *FuncArgsCollision) EXPECT() *FuncArgsCollision_Expecter {
	return &FuncArgsCollision_Expecter{mock: &_m.Mock}
}

 

// Foo provides a mock function for the type FuncArgsCollision
func (_mock *FuncArgsCollision) Foo(ret interface{}) error {  
	ret1 := _mock.Called(ret)

	if len(ret1) == 0 {
		panic("no return value specified for Foo")
	}

		
	var r0 error 
	if returnFunc, ok := ret1.Get(0).(func(interface{}) error); ok {
		r0 = returnFunc(ret)
	} else {
		r0 = ret1.Error(0)
	} 
	return r0
}



// FuncArgsCollision_Foo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Foo'
type FuncArgsCollision_Foo_Call struct {
	*mock.Call
}



// Foo is a helper method to define mock.On call
//  - ret
func (_e *FuncArgsCollision_Expecter) Foo(ret interface{}, ) *FuncArgsCollision_Foo_Call {
	return &FuncArgsCollision_Foo_Call{Call: _e.mock.On("Foo",ret, )}
}

func (_c *FuncArgsCollision_Foo_Call) Run(run func(ret interface{})) *FuncArgsCollision_Foo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}),)
	})
	return _c
}

func (_c *FuncArgsCollision_Foo_Call) Return(err error) *FuncArgsCollision_Foo_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *FuncArgsCollision_Foo_Call) RunAndReturn(run func(ret interface{})error) *FuncArgsCollision_Foo_Call {
	_c.Call.Return(run)
	return _c
}
  
