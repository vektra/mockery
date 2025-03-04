// Code generated by mockery. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// MapToInterface is an autogenerated mock type for the MapToInterface type
type MapToInterface struct {
	mock.Mock
}

type MapToInterface_Expecter struct {
	mock *mock.Mock
}

func (_m *MapToInterface) EXPECT() *MapToInterface_Expecter {
	return &MapToInterface_Expecter{mock: &_m.Mock}
}

// Foo provides a mock function with given fields: arg1
func (_m *MapToInterface) Foo(arg1 ...map[string]interface{}) {
	_va := make([]interface{}, len(arg1))
	for _i := range arg1 {
		_va[_i] = arg1[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	_m.Called(_ca...)
}

// MapToInterface_Foo_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Foo'
type MapToInterface_Foo_Call struct {
	*mock.Call
}

// Foo is a helper method to define mock.On call
//   - arg1 ...map[string]interface{}
func (_e *MapToInterface_Expecter) Foo(arg1 ...interface{}) *MapToInterface_Foo_Call {
	return &MapToInterface_Foo_Call{Call: _e.mock.On("Foo",
		append([]interface{}{}, arg1...)...)}
}

func (_c *MapToInterface_Foo_Call) Run(run func(arg1 ...map[string]interface{})) *MapToInterface_Foo_Call {
	_c.Call.Run(func(args mock.Arguments) {
		variadicArgs := make([]map[string]interface{}, len(args)-0)
		for i, a := range args[0:] {
			if a != nil {
				variadicArgs[i] = a.(map[string]interface{})
			}
		}
		run(variadicArgs...)
	})
	return _c
}

func (_c *MapToInterface_Foo_Call) Return() *MapToInterface_Foo_Call {
	_c.Call.Return()
	return _c
}

func (_c *MapToInterface_Foo_Call) RunAndReturn(run func(...map[string]interface{})) *MapToInterface_Foo_Call {
	_c.Run(run)
	return _c
}

// NewMapToInterface creates a new instance of MapToInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMapToInterface(t interface {
	mock.TestingT
	Cleanup(func())
}) *MapToInterface {
	mock := &MapToInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
