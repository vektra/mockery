// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// VariadicMock is a mock implementation of Variadic.
//
//	func TestSomethingThatUsesVariadic(t *testing.T) {
//
//		// make and configure a mocked Variadic
//		mockedVariadic := &VariadicMock{
//			VariadicFunctionFunc: func(str string, vFunc func(args1 string, args2 ...interface{}) interface{}) error {
//				panic("mock out the VariadicFunction method")
//			},
//		}
//
//		// use mockedVariadic in code that requires Variadic
//		// and then make assertions.
//
//	}
type VariadicMock struct {
	// VariadicFunctionFunc mocks the VariadicFunction method.
	VariadicFunctionFunc func(str string, vFunc func(args1 string, args2 ...interface{}) interface{}) error

	// calls tracks calls to the methods.
	calls struct {
		// VariadicFunction holds details about calls to the VariadicFunction method.
		VariadicFunction []struct {
			// Str is the str argument value.
			Str string
			// VFunc is the vFunc argument value.
			VFunc func(args1 string, args2 ...interface{}) interface{}
		}
	}
	lockVariadicFunction sync.RWMutex
}

// VariadicFunction calls VariadicFunctionFunc.
func (mock *VariadicMock) VariadicFunction(str string, vFunc func(args1 string, args2 ...interface{}) interface{}) error {
	if mock.VariadicFunctionFunc == nil {
		panic("VariadicMock.VariadicFunctionFunc: method is nil but Variadic.VariadicFunction was just called")
	}
	callInfo := struct {
		Str   string
		VFunc func(args1 string, args2 ...interface{}) interface{}
	}{
		Str:   str,
		VFunc: vFunc,
	}
	mock.lockVariadicFunction.Lock()
	mock.calls.VariadicFunction = append(mock.calls.VariadicFunction, callInfo)
	mock.lockVariadicFunction.Unlock()
	return mock.VariadicFunctionFunc(str, vFunc)
}

// VariadicFunctionCalls gets all the calls that were made to VariadicFunction.
// Check the length with:
//
//	len(mockedVariadic.VariadicFunctionCalls())
func (mock *VariadicMock) VariadicFunctionCalls() []struct {
	Str   string
	VFunc func(args1 string, args2 ...interface{}) interface{}
} {
	var calls []struct {
		Str   string
		VFunc func(args1 string, args2 ...interface{}) interface{}
	}
	mock.lockVariadicFunction.RLock()
	calls = mock.calls.VariadicFunction
	mock.lockVariadicFunction.RUnlock()
	return calls
}

// ResetVariadicFunctionCalls reset all the calls that were made to VariadicFunction.
func (mock *VariadicMock) ResetVariadicFunctionCalls() {
	mock.lockVariadicFunction.Lock()
	mock.calls.VariadicFunction = nil
	mock.lockVariadicFunction.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *VariadicMock) ResetCalls() {
	mock.lockVariadicFunction.Lock()
	mock.calls.VariadicFunction = nil
	mock.lockVariadicFunction.Unlock()
}
