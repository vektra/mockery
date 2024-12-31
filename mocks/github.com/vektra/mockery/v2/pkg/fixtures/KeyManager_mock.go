// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	test "github.com/vektra/mockery/v2/pkg/fixtures"
)

// NewKeyManager creates a new instance of KeyManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewKeyManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *KeyManager {
	mock := &KeyManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// KeyManager is an autogenerated mock type for the KeyManager type
type KeyManager struct {
	mock.Mock
}

type KeyManager_Expecter struct {
	mock *mock.Mock
}

func (_m *KeyManager) EXPECT() *KeyManager_Expecter {
	return &KeyManager_Expecter{mock: &_m.Mock}
}

// GetKey provides a mock function for the type KeyManager
func (_mock *KeyManager) GetKey(sParam string, vParam uint16) ([]byte, *test.Err) {
	ret := _mock.Called(sParam, vParam)

	if len(ret) == 0 {
		panic("no return value specified for GetKey")
	}

	var r0 []byte
	var r1 *test.Err
	if returnFunc, ok := ret.Get(0).(func(string, uint16) ([]byte, *test.Err)); ok {
		return returnFunc(sParam, vParam)
	}
	if returnFunc, ok := ret.Get(0).(func(string, uint16) []byte); ok {
		r0 = returnFunc(sParam, vParam)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(string, uint16) *test.Err); ok {
		r1 = returnFunc(sParam, vParam)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*test.Err)
		}
	}
	return r0, r1
}

// KeyManager_GetKey_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetKey'
type KeyManager_GetKey_Call struct {
	*mock.Call
}

// GetKey is a helper method to define mock.On call
//   - sParam
//   - vParam
func (_e *KeyManager_Expecter) GetKey(sParam interface{}, vParam interface{}) *KeyManager_GetKey_Call {
	return &KeyManager_GetKey_Call{Call: _e.mock.On("GetKey", sParam, vParam)}
}

func (_c *KeyManager_GetKey_Call) Run(run func(sParam string, vParam uint16)) *KeyManager_GetKey_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(uint16))
	})
	return _c
}

func (_c *KeyManager_GetKey_Call) Return(bytesOutParam []byte, errOutParam *test.Err) *KeyManager_GetKey_Call {
	_c.Call.Return(bytesOutParam, errOutParam)
	return _c
}

func (_c *KeyManager_GetKey_Call) RunAndReturn(run func(sParam string, vParam uint16) ([]byte, *test.Err)) *KeyManager_GetKey_Call {
	_c.Call.Return(run)
	return _c
}
