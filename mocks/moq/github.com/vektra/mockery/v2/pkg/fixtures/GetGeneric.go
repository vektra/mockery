// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package testfoo

import (
	"sync"

	test "github.com/vektra/mockery/v2/pkg/fixtures"
	"github.com/vektra/mockery/v2/pkg/fixtures/constraints"
)

// Ensure, that GetGeneric does implement test.GetGeneric.
// If this is not the case, regenerate this file with moq.
var _ test.GetGeneric[int] = &GetGeneric[int]{}

// GetGeneric is a mock implementation of test.GetGeneric.
//
//	func TestSomethingThatUsesGetGeneric(t *testing.T) {
//
//		// make and configure a mocked test.GetGeneric
//		mockedGetGeneric := &GetGeneric{
//			GetFunc: func() T {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedGetGeneric in code that requires test.GetGeneric
//		// and then make assertions.
//
//	}
type GetGeneric[T constraints.Integer] struct {
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
func (mock *GetGeneric[T]) Get() T {
	if mock.GetFunc == nil {
		panic("GetGeneric.GetFunc: method is nil but GetGeneric.Get was just called")
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
//	len(mockedGetGeneric.GetCalls())
func (mock *GetGeneric[T]) GetCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}
