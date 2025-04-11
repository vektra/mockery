// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery
// template: testify
// TEST MOCKERY BOILERPLATE

package same_name_arg_and_type

import (
	mock "github.com/stretchr/testify/mock"
)

// newmockinterfaceA creates a new instance of mockinterfaceA. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newmockinterfaceA(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockinterfaceA {
	mock := &mockinterfaceA{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// mockinterfaceA is an autogenerated mock type for the interfaceA type
type mockinterfaceA struct {
	mock.Mock
}

type mockinterfaceA_Expecter struct {
	mock *mock.Mock
}

func (_m *mockinterfaceA) EXPECT() *mockinterfaceA_Expecter {
	return &mockinterfaceA_Expecter{mock: &_m.Mock}
}

// DoB provides a mock function for the type mockinterfaceA
func (_mock *mockinterfaceA) DoB(interfaceB1 interfaceB) interfaceB {
	ret := _mock.Called(interfaceB1)

	if len(ret) == 0 {
		panic("no return value specified for DoB")
	}

	var r0 interfaceB
	if returnFunc, ok := ret.Get(0).(func(interfaceB) interfaceB); ok {
		r0 = returnFunc(interfaceB1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaceB)
		}
	}
	return r0
}

// mockinterfaceA_DoB_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DoB'
type mockinterfaceA_DoB_Call struct {
	*mock.Call
}

// DoB is a helper method to define mock.On call
//   - interfaceB1
func (_e *mockinterfaceA_Expecter) DoB(interfaceB1 interface{}) *mockinterfaceA_DoB_Call {
	return &mockinterfaceA_DoB_Call{Call: _e.mock.On("DoB", interfaceB1)}
}

func (_c *mockinterfaceA_DoB_Call) Run(run func(interfaceB1 interfaceB)) *mockinterfaceA_DoB_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interfaceB))
	})
	return _c
}

func (_c *mockinterfaceA_DoB_Call) Return(interfaceBMoqParam interfaceB) *mockinterfaceA_DoB_Call {
	_c.Call.Return(interfaceBMoqParam)
	return _c
}

func (_c *mockinterfaceA_DoB_Call) RunAndReturn(run func(interfaceB1 interfaceB) interfaceB) *mockinterfaceA_DoB_Call {
	_c.Call.Return(run)
	return _c
}

// DoB0 provides a mock function for the type mockinterfaceA
func (_mock *mockinterfaceA) DoB0(interfaceB interfaceB0) interfaceB0 {
	ret := _mock.Called(interfaceB)

	if len(ret) == 0 {
		panic("no return value specified for DoB0")
	}

	var r0 interfaceB0
	if returnFunc, ok := ret.Get(0).(func(interfaceB0) interfaceB0); ok {
		r0 = returnFunc(interfaceB)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaceB0)
		}
	}
	return r0
}

// mockinterfaceA_DoB0_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DoB0'
type mockinterfaceA_DoB0_Call struct {
	*mock.Call
}

// DoB0 is a helper method to define mock.On call
//   - interfaceB
func (_e *mockinterfaceA_Expecter) DoB0(interfaceB interface{}) *mockinterfaceA_DoB0_Call {
	return &mockinterfaceA_DoB0_Call{Call: _e.mock.On("DoB0", interfaceB)}
}

func (_c *mockinterfaceA_DoB0_Call) Run(run func(interfaceB interfaceB0)) *mockinterfaceA_DoB0_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interfaceB0))
	})
	return _c
}

func (_c *mockinterfaceA_DoB0_Call) Return(interfaceB0MoqParam interfaceB0) *mockinterfaceA_DoB0_Call {
	_c.Call.Return(interfaceB0MoqParam)
	return _c
}

func (_c *mockinterfaceA_DoB0_Call) RunAndReturn(run func(interfaceB interfaceB0) interfaceB0) *mockinterfaceA_DoB0_Call {
	_c.Call.Return(run)
	return _c
}

// DoB0v2 provides a mock function for the type mockinterfaceA
func (_mock *mockinterfaceA) DoB0v2(interfaceB01 interfaceB0) interfaceB0 {
	ret := _mock.Called(interfaceB01)

	if len(ret) == 0 {
		panic("no return value specified for DoB0v2")
	}

	var r0 interfaceB0
	if returnFunc, ok := ret.Get(0).(func(interfaceB0) interfaceB0); ok {
		r0 = returnFunc(interfaceB01)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaceB0)
		}
	}
	return r0
}

// mockinterfaceA_DoB0v2_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DoB0v2'
type mockinterfaceA_DoB0v2_Call struct {
	*mock.Call
}

// DoB0v2 is a helper method to define mock.On call
//   - interfaceB01
func (_e *mockinterfaceA_Expecter) DoB0v2(interfaceB01 interface{}) *mockinterfaceA_DoB0v2_Call {
	return &mockinterfaceA_DoB0v2_Call{Call: _e.mock.On("DoB0v2", interfaceB01)}
}

func (_c *mockinterfaceA_DoB0v2_Call) Run(run func(interfaceB01 interfaceB0)) *mockinterfaceA_DoB0v2_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interfaceB0))
	})
	return _c
}

func (_c *mockinterfaceA_DoB0v2_Call) Return(interfaceB0MoqParam interfaceB0) *mockinterfaceA_DoB0v2_Call {
	_c.Call.Return(interfaceB0MoqParam)
	return _c
}

func (_c *mockinterfaceA_DoB0v2_Call) RunAndReturn(run func(interfaceB01 interfaceB0) interfaceB0) *mockinterfaceA_DoB0v2_Call {
	_c.Call.Return(run)
	return _c
}

// newmockinterfaceB creates a new instance of mockinterfaceB. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newmockinterfaceB(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockinterfaceB {
	mock := &mockinterfaceB{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// mockinterfaceB is an autogenerated mock type for the interfaceB type
type mockinterfaceB struct {
	mock.Mock
}

type mockinterfaceB_Expecter struct {
	mock *mock.Mock
}

func (_m *mockinterfaceB) EXPECT() *mockinterfaceB_Expecter {
	return &mockinterfaceB_Expecter{mock: &_m.Mock}
}

// GetData provides a mock function for the type mockinterfaceB
func (_mock *mockinterfaceB) GetData() int {
	ret := _mock.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetData")
	}

	var r0 int
	if returnFunc, ok := ret.Get(0).(func() int); ok {
		r0 = returnFunc()
	} else {
		r0 = ret.Get(0).(int)
	}
	return r0
}

// mockinterfaceB_GetData_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetData'
type mockinterfaceB_GetData_Call struct {
	*mock.Call
}

// GetData is a helper method to define mock.On call
func (_e *mockinterfaceB_Expecter) GetData() *mockinterfaceB_GetData_Call {
	return &mockinterfaceB_GetData_Call{Call: _e.mock.On("GetData")}
}

func (_c *mockinterfaceB_GetData_Call) Run(run func()) *mockinterfaceB_GetData_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockinterfaceB_GetData_Call) Return(n int) *mockinterfaceB_GetData_Call {
	_c.Call.Return(n)
	return _c
}

func (_c *mockinterfaceB_GetData_Call) RunAndReturn(run func() int) *mockinterfaceB_GetData_Call {
	_c.Call.Return(run)
	return _c
}

// newmockinterfaceB0 creates a new instance of mockinterfaceB0. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newmockinterfaceB0(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockinterfaceB0 {
	mock := &mockinterfaceB0{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// mockinterfaceB0 is an autogenerated mock type for the interfaceB0 type
type mockinterfaceB0 struct {
	mock.Mock
}

type mockinterfaceB0_Expecter struct {
	mock *mock.Mock
}

func (_m *mockinterfaceB0) EXPECT() *mockinterfaceB0_Expecter {
	return &mockinterfaceB0_Expecter{mock: &_m.Mock}
}

// DoB0 provides a mock function for the type mockinterfaceB0
func (_mock *mockinterfaceB0) DoB0(interfaceB01 interfaceB0) interfaceB0 {
	ret := _mock.Called(interfaceB01)

	if len(ret) == 0 {
		panic("no return value specified for DoB0")
	}

	var r0 interfaceB0
	if returnFunc, ok := ret.Get(0).(func(interfaceB0) interfaceB0); ok {
		r0 = returnFunc(interfaceB01)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interfaceB0)
		}
	}
	return r0
}

// mockinterfaceB0_DoB0_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DoB0'
type mockinterfaceB0_DoB0_Call struct {
	*mock.Call
}

// DoB0 is a helper method to define mock.On call
//   - interfaceB01
func (_e *mockinterfaceB0_Expecter) DoB0(interfaceB01 interface{}) *mockinterfaceB0_DoB0_Call {
	return &mockinterfaceB0_DoB0_Call{Call: _e.mock.On("DoB0", interfaceB01)}
}

func (_c *mockinterfaceB0_DoB0_Call) Run(run func(interfaceB01 interfaceB0)) *mockinterfaceB0_DoB0_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(interfaceB0))
	})
	return _c
}

func (_c *mockinterfaceB0_DoB0_Call) Return(interfaceB0MoqParam interfaceB0) *mockinterfaceB0_DoB0_Call {
	_c.Call.Return(interfaceB0MoqParam)
	return _c
}

func (_c *mockinterfaceB0_DoB0_Call) RunAndReturn(run func(interfaceB01 interfaceB0) interfaceB0) *mockinterfaceB0_DoB0_Call {
	_c.Call.Return(run)
	return _c
}
