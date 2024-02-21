// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"

	test "github.com/vektra/mockery/v2/pkg/fixtures"
)

// Ensure, that Requester2Moq does implement test.Requester2.
// If this is not the case, regenerate this file with moq.
var _ test.Requester2 = &Requester2Moq{}

// Requester2Moq is a mock implementation of test.Requester2.
//
//	func TestSomethingThatUsesRequester2(t *testing.T) {
//
//		// make and configure a mocked test.Requester2
//		mockedRequester2 := &Requester2Moq{
//			GetFunc: func(path string) error {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedRequester2 in code that requires test.Requester2
//		// and then make assertions.
//
//	}
type Requester2Moq struct {
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
func (mock *Requester2Moq) Get(path string) error {
	if mock.GetFunc == nil {
		panic("Requester2Moq.GetFunc: method is nil but Requester2.Get was just called")
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
func (mock *Requester2Moq) GetCalls() []struct {
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
func (mock *Requester2Moq) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *Requester2Moq) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}
