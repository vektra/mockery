// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
	"net/http"

	mock "github.com/stretchr/testify/mock"
	number_dir_http "github.com/vektra/mockery/v2/pkg/fixtures/12345678/http"
	my_http "github.com/vektra/mockery/v2/pkg/fixtures/http"
)

// NewExample creates a new instance of Example. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewExample(t interface {
	mock.TestingT
	Cleanup(func())
}) *Example {
	mock := &Example{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// Example is an autogenerated mock type for the Example type
type Example struct {
	mock.Mock
}

type Example_Expecter struct {
	mock *mock.Mock
}

func (_m *Example) EXPECT() *Example_Expecter {
	return &Example_Expecter{mock: &_m.Mock}
}

// A provides a mock function for the type Example
func (_mock *Example) A() http.Flusher {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for A")
	}

	var r0 http.Flusher
	if returnFunc, ok := ret.Get(0).(func() http.Flusher); ok {
		r0 = returnFunc()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Flusher)
		}
	}
	return r0
}

// Example_A_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'A'
type Example_A_Call struct {
	*mock.Call
}

// A is a helper method to define mock.On call
func (_e *Example_Expecter) A() *Example_A_Call {
	return &Example_A_Call{Call: _e.mock.On("A")}
}

func (_c *Example_A_Call) Run(run func()) *Example_A_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Example_A_Call) Return(flusherOutParam http.Flusher) *Example_A_Call {
	_c.Call.Return(flusherOutParam)
	return _c
}

func (_c *Example_A_Call) RunAndReturn(run func() http.Flusher) *Example_A_Call {
	_c.Call.Return(run)
	return _c
}

// B provides a mock function for the type Example
func (_mock *Example) B(fixtureshttpParam string) my_http.MyStruct {
	ret := _mock.Called(fixtureshttpParam)

	if len(ret) == 0 {
		panic("no return value specified for B")
	}

	var r0 my_http.MyStruct
	if returnFunc, ok := ret.Get(0).(func(string) my_http.MyStruct); ok {
		r0 = returnFunc(fixtureshttpParam)
	} else {
		r0 = ret.Get(0).(my_http.MyStruct)
	}
	return r0
}

// Example_B_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'B'
type Example_B_Call struct {
	*mock.Call
}

// B is a helper method to define mock.On call
//   - fixtureshttpParam
func (_e *Example_Expecter) B(fixtureshttpParam interface{}) *Example_B_Call {
	return &Example_B_Call{Call: _e.mock.On("B", fixtureshttpParam)}
}

func (_c *Example_B_Call) Run(run func(fixtureshttpParam string)) *Example_B_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Example_B_Call) Return(myStructOutParam my_http.MyStruct) *Example_B_Call {
	_c.Call.Return(myStructOutParam)
	return _c
}

func (_c *Example_B_Call) RunAndReturn(run func(fixtureshttpParam string) my_http.MyStruct) *Example_B_Call {
	_c.Call.Return(run)
	return _c
}

// C provides a mock function for the type Example
func (_mock *Example) C(fixtureshttpParam string) number_dir_http.MyStruct {
	ret := _mock.Called(fixtureshttpParam)

	if len(ret) == 0 {
		panic("no return value specified for C")
	}

	var r0 number_dir_http.MyStruct
	if returnFunc, ok := ret.Get(0).(func(string) number_dir_http.MyStruct); ok {
		r0 = returnFunc(fixtureshttpParam)
	} else {
		r0 = ret.Get(0).(number_dir_http.MyStruct)
	}
	return r0
}

// Example_C_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'C'
type Example_C_Call struct {
	*mock.Call
}

// C is a helper method to define mock.On call
//   - fixtureshttpParam
func (_e *Example_Expecter) C(fixtureshttpParam interface{}) *Example_C_Call {
	return &Example_C_Call{Call: _e.mock.On("C", fixtureshttpParam)}
}

func (_c *Example_C_Call) Run(run func(fixtureshttpParam string)) *Example_C_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *Example_C_Call) Return(myStructOutParam number_dir_http.MyStruct) *Example_C_Call {
	_c.Call.Return(myStructOutParam)
	return _c
}

func (_c *Example_C_Call) RunAndReturn(run func(fixtureshttpParam string) number_dir_http.MyStruct) *Example_C_Call {
	_c.Call.Return(run)
	return _c
}
