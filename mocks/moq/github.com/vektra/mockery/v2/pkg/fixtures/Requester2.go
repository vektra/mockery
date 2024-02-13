// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// Requester2 is a mock implementation of Requester2.
//
//	func TestSomethingThatUsesRequester2(t *testing.T) {
//
//		// make and configure a mocked Requester2
//		mockedRequester2 := &Requester2{
//			GetFunc: func(path string) error {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedRequester2 in code that requires Requester2
//		// and then make assertions.
//
//	}
type Requester2 struct {
	// GetFunc mocks the Get method.
	GetFunc func(path string) error

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
			// Path is the path argument value.
			Path string
		}
	}
	lockGet sync.RWMutex
}

// Get calls GetFunc.
func (mock *Requester2) Get(path string) error {
	if mock.GetFunc == nil {
		panic("Requester2.GetFunc: method is nil but Requester2.Get was just called")
	}
	callInfo := struct {
		Path string
	}{
		Path: path,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(path)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//
//	len(mockedRequester2.GetCalls())
func (mock *Requester2) GetCalls() []struct {
	Path string
} {
	var calls []struct {
		Path string
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// ResetGetCalls reset all the calls that were made to Get.
func (mock *Requester2) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *Requester2) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}
