// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"github.com/vektra/mockery/v2/pkg/fixtures/constraints"
	"sync"
)

// EmbeddedGetMock is a mock implementation of test.EmbeddedGet.
//
//	func TestSomethingThatUsesEmbeddedGet(t *testing.T) {
//
//		// make and configure a mocked test.EmbeddedGet
//		mockedEmbeddedGet := &EmbeddedGetMock{
//			GetFunc: func() T {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedEmbeddedGet in code that requires test.EmbeddedGet
//		// and then make assertions.
//
//	}
type EmbeddedGetMock[T constraints.Signed] struct {
	// GetFunc mocks the Get method.
	GetFunc func() T

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
		}
	}
	lockGet sync.RWMutex
}

// Get calls GetFunc.
func (mock *EmbeddedGetMock[T]) Get() T {
	if mock.GetFunc == nil {
		panic("EmbeddedGetMock.GetFunc: method is nil but EmbeddedGet.Get was just called")
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
//	len(mockedEmbeddedGet.GetCalls())
func (mock *EmbeddedGetMock[T]) GetCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// ResetGetCalls reset all the calls that were made to Get.
func (mock *EmbeddedGetMock[T]) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *EmbeddedGetMock[T]) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}
