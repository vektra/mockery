// Code generated by mockery. DO NOT EDIT.

package iface_typed_param_test

import (
	io "io"

	mock "github.com/stretchr/testify/mock"
)

// MockGetterIfaceTypedParam is an autogenerated mock type for the GetterIfaceTypedParam type
type MockGetterIfaceTypedParam[T io.Reader] struct {
	mock.Mock
}

type MockGetterIfaceTypedParam_Expecter[T io.Reader] struct {
	mock *mock.Mock
}

func (_m *MockGetterIfaceTypedParam[T]) EXPECT() *MockGetterIfaceTypedParam_Expecter[T] {
	return &MockGetterIfaceTypedParam_Expecter[T]{mock: &_m.Mock}
}

// Get provides a mock function with no fields
func (_m *MockGetterIfaceTypedParam[T]) Get() T {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 T
	if rf, ok := ret.Get(0).(func() T); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(T)
		}
	}

	return r0
}

// MockGetterIfaceTypedParam_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type MockGetterIfaceTypedParam_Get_Call[T io.Reader] struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
func (_e *MockGetterIfaceTypedParam_Expecter[T]) Get() *MockGetterIfaceTypedParam_Get_Call[T] {
	return &MockGetterIfaceTypedParam_Get_Call[T]{Call: _e.mock.On("Get")}
}

func (_c *MockGetterIfaceTypedParam_Get_Call[T]) Run(run func()) *MockGetterIfaceTypedParam_Get_Call[T] {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockGetterIfaceTypedParam_Get_Call[T]) Return(_a0 T) *MockGetterIfaceTypedParam_Get_Call[T] {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockGetterIfaceTypedParam_Get_Call[T]) RunAndReturn(run func() T) *MockGetterIfaceTypedParam_Get_Call[T] {
	_c.Call.Return(run)
	return _c
}

// NewMockGetterIfaceTypedParam creates a new instance of MockGetterIfaceTypedParam. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockGetterIfaceTypedParam[T io.Reader](t interface {
	mock.TestingT
	Cleanup(func())
}) *MockGetterIfaceTypedParam[T] {
	mock := &MockGetterIfaceTypedParam[T]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
