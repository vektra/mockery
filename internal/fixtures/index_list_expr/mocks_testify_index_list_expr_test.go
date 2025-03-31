// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify
// TEST MOCKERY BOILERPLATE

package index_list_expr

import (
	mock "github.com/stretchr/testify/mock"
)

// NewMockGenericMultipleTypes creates a new instance of MockGenericMultipleTypes. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockGenericMultipleTypes[T1 any, T2 any, T3 any](t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGenericMultipleTypes[T1, T2, T3] {
	mock := &MockGenericMultipleTypes[T1, T2, T3]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockGenericMultipleTypes is an autogenerated mock type for the GenericMultipleTypes type
type MockGenericMultipleTypes[T1 any, T2 any, T3 any] struct {
	mock.Mock
}

type MockGenericMultipleTypes_Expecter[T1 any, T2 any, T3 any] struct {
	mock *mock.Mock
}

func (_m *MockGenericMultipleTypes[T1, T2, T3]) EXPECT() *MockGenericMultipleTypes_Expecter[T1, T2, T3] {
	return &MockGenericMultipleTypes_Expecter[T1, T2, T3]{mock: &_m.Mock}
}

// Func provides a mock function for the type MockGenericMultipleTypes
func (_mock *MockGenericMultipleTypes[T1, T2, T3]) Func(arg1 *T1, arg2 T2) T3 {
	ret := _mock.Called(arg1, arg2)

	if len(ret) == 0 {
		panic("no return value specified for Func")
	}

	var r0 T3
	if returnFunc, ok := ret.Get(0).(func(*T1, T2) T3); ok {
		r0 = returnFunc(arg1, arg2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(T3)
		}
	}
	return r0
}

// MockGenericMultipleTypes_Func_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Func'
type MockGenericMultipleTypes_Func_Call[T1 any, T2 any, T3 any] struct {
	*mock.Call
}

// Func is a helper method to define mock.On call
//   - arg1
//   - arg2
func (_e *MockGenericMultipleTypes_Expecter[T1, T2, T3]) Func(arg1 interface{}, arg2 interface{}) *MockGenericMultipleTypes_Func_Call[T1, T2, T3] {
	return &MockGenericMultipleTypes_Func_Call[T1, T2, T3]{Call: _e.mock.On("Func", arg1, arg2)}
}

func (_c *MockGenericMultipleTypes_Func_Call[T1, T2, T3]) Run(run func(arg1 *T1, arg2 T2)) *MockGenericMultipleTypes_Func_Call[T1, T2, T3] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*T1), args[1].(T2))
	})
	return _c
}

func (_c *MockGenericMultipleTypes_Func_Call[T1, T2, T3]) Return(v T3) *MockGenericMultipleTypes_Func_Call[T1, T2, T3] {
	_c.Call.Return(v)
	return _c
}

func (_c *MockGenericMultipleTypes_Func_Call[T1, T2, T3]) RunAndReturn(run func(arg1 *T1, arg2 T2) T3) *MockGenericMultipleTypes_Func_Call[T1, T2, T3] {
	_c.Call.Return(run)
	return _c
}

// NewMockIndexListExpr creates a new instance of MockIndexListExpr. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockIndexListExpr(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockIndexListExpr {
	mock := &MockIndexListExpr{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// MockIndexListExpr is an autogenerated mock type for the IndexListExpr type
type MockIndexListExpr struct {
	mock.Mock
}

type MockIndexListExpr_Expecter struct {
	mock *mock.Mock
}

func (_m *MockIndexListExpr) EXPECT() *MockIndexListExpr_Expecter {
	return &MockIndexListExpr_Expecter{mock: &_m.Mock}
}

// Func provides a mock function for the type MockIndexListExpr
func (_mock *MockIndexListExpr) Func(arg1 *int, arg2 string) bool {
	ret := _mock.Called(arg1, arg2)

	if len(ret) == 0 {
		panic("no return value specified for Func")
	}

	var r0 bool
	if returnFunc, ok := ret.Get(0).(func(*int, string) bool); ok {
		r0 = returnFunc(arg1, arg2)
	} else {
		r0 = ret.Get(0).(bool)
	}
	return r0
}

// MockIndexListExpr_Func_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Func'
type MockIndexListExpr_Func_Call struct {
	*mock.Call
}

// Func is a helper method to define mock.On call
//   - arg1
//   - arg2
func (_e *MockIndexListExpr_Expecter) Func(arg1 interface{}, arg2 interface{}) *MockIndexListExpr_Func_Call {
	return &MockIndexListExpr_Func_Call{Call: _e.mock.On("Func", arg1, arg2)}
}

func (_c *MockIndexListExpr_Func_Call) Run(run func(arg1 *int, arg2 string)) *MockIndexListExpr_Func_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*int), args[1].(string))
	})
	return _c
}

func (_c *MockIndexListExpr_Func_Call) Return(b bool) *MockIndexListExpr_Func_Call {
	_c.Call.Return(b)
	return _c
}

func (_c *MockIndexListExpr_Func_Call) RunAndReturn(run func(arg1 *int, arg2 string) bool) *MockIndexListExpr_Func_Call {
	_c.Call.Return(run)
	return _c
}
