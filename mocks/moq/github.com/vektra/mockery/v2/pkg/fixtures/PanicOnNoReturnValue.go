// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// PanicOnNoReturnValueMock is a mock implementation of PanicOnNoReturnValue.
//
//	func TestSomethingThatUsesPanicOnNoReturnValue(t *testing.T) {
//
//		// make and configure a mocked PanicOnNoReturnValue
//		mockedPanicOnNoReturnValue := &PanicOnNoReturnValueMock{
//			DoSomethingFunc: func() string {
//				panic("mock out the DoSomething method")
//			},
//		}
//
//		// use mockedPanicOnNoReturnValue in code that requires PanicOnNoReturnValue
//		// and then make assertions.
//
//	}
type PanicOnNoReturnValueMock struct {
	// DoSomethingFunc mocks the DoSomething method.
	DoSomethingFunc func() string

	// calls tracks calls to the methods.
	calls struct {
		// DoSomething holds details about calls to the DoSomething method.
		DoSomething []struct {
		}
	}
	lockDoSomething sync.RWMutex
}

// DoSomething calls DoSomethingFunc.
func (mock *PanicOnNoReturnValueMock) DoSomething() string {
	if mock.DoSomethingFunc == nil {
		panic("PanicOnNoReturnValueMock.DoSomethingFunc: method is nil but PanicOnNoReturnValue.DoSomething was just called")
	}
	callInfo := struct {
	}{}
	mock.lockDoSomething.Lock()
	mock.calls.DoSomething = append(mock.calls.DoSomething, callInfo)
	mock.lockDoSomething.Unlock()
	return mock.DoSomethingFunc()
}

// DoSomethingCalls gets all the calls that were made to DoSomething.
// Check the length with:
//
//	len(mockedPanicOnNoReturnValue.DoSomethingCalls())
func (mock *PanicOnNoReturnValueMock) DoSomethingCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockDoSomething.RLock()
	calls = mock.calls.DoSomething
	mock.lockDoSomething.RUnlock()
	return calls
}

// ResetDoSomethingCalls reset all the calls that were made to DoSomething.
func (mock *PanicOnNoReturnValueMock) ResetDoSomethingCalls() {
	mock.lockDoSomething.Lock()
	mock.calls.DoSomething = nil
	mock.lockDoSomething.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *PanicOnNoReturnValueMock) ResetCalls() {
	mock.lockDoSomething.Lock()
	mock.calls.DoSomething = nil
	mock.lockDoSomething.Unlock()
}
