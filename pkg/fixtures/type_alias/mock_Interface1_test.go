
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package type_alias_test

import (
	"github.com/vektra/mockery/v3/pkg/fixtures/type_alias"
    mock "github.com/stretchr/testify/mock"
)

 
// NewInterface1 creates a new instance of Interface1. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewInterface1 (t interface {
	mock.TestingT
	Cleanup(func())
}) *Interface1 {
	mock := &Interface1{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}


// Interface1 is an autogenerated mock type for the Interface1 type
type Interface1 struct {
	mock.Mock
}

type Interface1_Expecter struct {
	mock *mock.Mock
}

func (_m *Interface1) EXPECT() *Interface1_Expecter {
	return &Interface1_Expecter{mock: &_m.Mock}
}

 

// Foo provides a mock function for the type Interface1
func (_mock *Interface1) Foo() type_alias.Type {  
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for Foo")
	}

		
	var r0 type_alias.Type 
	if returnFunc, ok := ret.Get(0).(func() type_alias.Type); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(type_alias.Type)
	} 
	return r0
}



// Interface1_Foo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Foo'
type Interface1_Foo_Call struct {
	*mock.Call
}



// Foo is a helper method to define mock.On call
func (_e *Interface1_Expecter) Foo() *Interface1_Foo_Call {
	return &Interface1_Foo_Call{Call: _e.mock.On("Foo", )}
}

func (_c *Interface1_Foo_Call) Run(run func()) *Interface1_Foo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Interface1_Foo_Call) Return(v type_alias.Type) *Interface1_Foo_Call {
	_c.Call.Return(v)
	return _c
}

func (_c *Interface1_Foo_Call) RunAndReturn(run func()type_alias.Type) *Interface1_Foo_Call {
	_c.Call.Return(run)
	return _c
}
  

