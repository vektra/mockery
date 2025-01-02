
// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
	test0 "github.com/vektra/mockery/v3/pkg/fixtures"
	"github.com/vektra/mockery/v3/pkg/fixtures/redefined_type_b"
    mock "github.com/stretchr/testify/mock"
)

 
// NewImportsSameAsPackage creates a new instance of ImportsSameAsPackage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewImportsSameAsPackage (t interface {
	mock.TestingT
	Cleanup(func())
}) *ImportsSameAsPackage {
	mock := &ImportsSameAsPackage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}


// ImportsSameAsPackage is an autogenerated mock type for the ImportsSameAsPackage type
type ImportsSameAsPackage struct {
	mock.Mock
}

type ImportsSameAsPackage_Expecter struct {
	mock *mock.Mock
}

func (_m *ImportsSameAsPackage) EXPECT() *ImportsSameAsPackage_Expecter {
	return &ImportsSameAsPackage_Expecter{mock: &_m.Mock}
}

 

// A provides a mock function for the type ImportsSameAsPackage
func (_mock *ImportsSameAsPackage) A() test.B {  
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for A")
	}

		
	var r0 test.B 
	if returnFunc, ok := ret.Get(0).(func() test.B); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(test.B)
	} 
	return r0
}



// ImportsSameAsPackage_A_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'A'
type ImportsSameAsPackage_A_Call struct {
	*mock.Call
}



// A is a helper method to define mock.On call
func (_e *ImportsSameAsPackage_Expecter) A() *ImportsSameAsPackage_A_Call {
	return &ImportsSameAsPackage_A_Call{Call: _e.mock.On("A", )}
}

func (_c *ImportsSameAsPackage_A_Call) Run(run func()) *ImportsSameAsPackage_A_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ImportsSameAsPackage_A_Call) Return(b test.B) *ImportsSameAsPackage_A_Call {
	_c.Call.Return(b)
	return _c
}

func (_c *ImportsSameAsPackage_A_Call) RunAndReturn(run func()test.B) *ImportsSameAsPackage_A_Call {
	_c.Call.Return(run)
	return _c
}
 

// B provides a mock function for the type ImportsSameAsPackage
func (_mock *ImportsSameAsPackage) B() test0.KeyManager {  
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for B")
	}

		
	var r0 test0.KeyManager 
	if returnFunc, ok := ret.Get(0).(func() test0.KeyManager); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(test0.KeyManager)
		}
	} 
	return r0
}



// ImportsSameAsPackage_B_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'B'
type ImportsSameAsPackage_B_Call struct {
	*mock.Call
}



// B is a helper method to define mock.On call
func (_e *ImportsSameAsPackage_Expecter) B() *ImportsSameAsPackage_B_Call {
	return &ImportsSameAsPackage_B_Call{Call: _e.mock.On("B", )}
}

func (_c *ImportsSameAsPackage_B_Call) Run(run func()) *ImportsSameAsPackage_B_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *ImportsSameAsPackage_B_Call) Return(keyManager test0.KeyManager) *ImportsSameAsPackage_B_Call {
	_c.Call.Return(keyManager)
	return _c
}

func (_c *ImportsSameAsPackage_B_Call) RunAndReturn(run func()test0.KeyManager) *ImportsSameAsPackage_B_Call {
	_c.Call.Return(run)
	return _c
}
 

// C provides a mock function for the type ImportsSameAsPackage
func (_mock *ImportsSameAsPackage) C(c test0.C)  {  _mock.Called(c)
	return 
}



// ImportsSameAsPackage_C_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'C'
type ImportsSameAsPackage_C_Call struct {
	*mock.Call
}



// C is a helper method to define mock.On call
//  - c
func (_e *ImportsSameAsPackage_Expecter) C(c interface{}, ) *ImportsSameAsPackage_C_Call {
	return &ImportsSameAsPackage_C_Call{Call: _e.mock.On("C",c, )}
}

func (_c *ImportsSameAsPackage_C_Call) Run(run func(c test0.C)) *ImportsSameAsPackage_C_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(test0.C),)
	})
	return _c
}

func (_c *ImportsSameAsPackage_C_Call) Return() *ImportsSameAsPackage_C_Call {
	_c.Call.Return()
	return _c
}

func (_c *ImportsSameAsPackage_C_Call) RunAndReturn(run func(c test0.C)) *ImportsSameAsPackage_C_Call {
	_c.Run(run)
	return _c
}
  
