
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
    mock "github.com/stretchr/testify/mock"
)

 
// NewBlank creates a new instance of Blank. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBlank (t interface {
	mock.TestingT
	Cleanup(func())
}) *Blank {
	mock := &Blank{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}


// Blank is an autogenerated mock type for the Blank type
type Blank struct {
	mock.Mock
}

type Blank_Expecter struct {
	mock *mock.Mock
}

func (_m *Blank) EXPECT() *Blank_Expecter {
	return &Blank_Expecter{mock: &_m.Mock}
}

 

// Create provides a mock function for the type Blank
func (_mock *Blank) Create(x interface{}) error {  
	ret := _mock.Called(x)

	if len(ret) == 0 {
		panic("no return value specified for Create")
	}

		
	var r0 error 
	if returnFunc, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = returnFunc(x)
	} else {
		r0 = ret.Error(0)
	} 
	return r0
}



// Blank_Create_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Create'
type Blank_Create_Call struct {
	*mock.Call
}



// Create is a helper method to define mock.On call
//  - x
func (_e *Blank_Expecter) Create(x interface{}, ) *Blank_Create_Call {
	return &Blank_Create_Call{Call: _e.mock.On("Create",x, )}
}

func (_c *Blank_Create_Call) Run(run func(x interface{})) *Blank_Create_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interface{}),)
	})
	return _c
}

func (_c *Blank_Create_Call) Return(err error) *Blank_Create_Call {
	_c.Call.Return(err)
	return _c
}

func (_c *Blank_Create_Call) RunAndReturn(run func(x interface{})error) *Blank_Create_Call {
	_c.Call.Return(run)
	return _c
}
  
