// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package testfoo

import (
	"sync"

	test "github.com/vektra/mockery/v2/pkg/fixtures"
)

// Ensure, that RequesterSlice does implement test.RequesterSlice.
// If this is not the case, regenerate this file with moq.
var _ test.RequesterSlice = &RequesterSlice{}

// RequesterSlice is a mock implementation of test.RequesterSlice.
//
//	func TestSomethingThatUsesRequesterSlice(t *testing.T) {
//
//		// make and configure a mocked test.RequesterSlice
//		mockedRequesterSlice := &RequesterSlice{
//			GetFunc: func(path string) ([]string, error) {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedRequesterSlice in code that requires test.RequesterSlice
//		// and then make assertions.
//
//	}
type RequesterSlice struct {
	// GetFunc mocks the Get method.
	GetFunc func(path string) ([]string, error)

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
func (mock *RequesterSlice) Get(path string) ([]string, error) {
	if mock.GetFunc == nil {
		panic("RequesterSlice.GetFunc: method is nil but RequesterSlice.Get was just called")
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
//	len(mockedRequesterSlice.GetCalls())
func (mock *RequesterSlice) GetCalls() []struct {
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
