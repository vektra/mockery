// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// GetInt is a mock implementation of GetInt.
//
//	func TestSomethingThatUsesGetInt(t *testing.T) {
//
//		// make and configure a mocked GetInt
//		mockedGetInt := &GetInt{
//			GetFunc: func() int {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedGetInt in code that requires GetInt
//		// and then make assertions.
//
//	}
type GetInt struct {
	// GetFunc mocks the Get method.
	GetFunc func() int

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
		}
	}
	lockGet sync.RWMutex
}

// Get calls GetFunc.
func (mock *GetInt) Get() int {
	if mock.GetFunc == nil {
		panic("GetInt.GetFunc: method is nil but GetInt.Get was just called")
	}
	callInfo := struct {
	}{}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc()
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//
//	len(mockedGetInt.GetCalls())
func (mock *GetInt) GetCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// ResetGetCalls reset all the calls that were made to Get.
func (mock *GetInt) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *GetInt) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}
