// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// Ensure, that RequesterArrayMoq does implement RequesterArray.
// If this is not the case, regenerate this file with moq.
var _ RequesterArray = &RequesterArrayMoq{}

// RequesterArrayMoq is a mock implementation of RequesterArray.
//
//	func TestSomethingThatUsesRequesterArray(t *testing.T) {
//
//		// make and configure a mocked RequesterArray
//		mockedRequesterArray := &RequesterArrayMoq{
//			GetFunc: func(path string) ([2]string, error) {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedRequesterArray in code that requires RequesterArray
//		// and then make assertions.
//
//	}
type RequesterArrayMoq struct {
	// GetFunc mocks the Get method.
	GetFunc func(path string) ([2]string, error)

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
func (mock *RequesterArrayMoq) Get(path string) ([2]string, error) {
	if mock.GetFunc == nil {
		panic("RequesterArrayMoq.GetFunc: method is nil but RequesterArray.Get was just called")
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
//	len(mockedRequesterArray.GetCalls())
func (mock *RequesterArrayMoq) GetCalls() []struct {
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
func (mock *RequesterArrayMoq) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *RequesterArrayMoq) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}