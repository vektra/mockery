// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// Ensure, that RequesterReturnElidedMoq does implement RequesterReturnElided.
// If this is not the case, regenerate this file with moq.
var _ RequesterReturnElided = &RequesterReturnElidedMoq{}

// RequesterReturnElidedMoq is a mock implementation of RequesterReturnElided.
//
//	func TestSomethingThatUsesRequesterReturnElided(t *testing.T) {
//
//		// make and configure a mocked RequesterReturnElided
//		mockedRequesterReturnElided := &RequesterReturnElidedMoq{
//			GetFunc: func(path string) (int, int, int, error) {
//				panic("mock out the Get method")
//			},
//			PutFunc: func(path string) (int, error) {
//				panic("mock out the Put method")
//			},
//		}
//
//		// use mockedRequesterReturnElided in code that requires RequesterReturnElided
//		// and then make assertions.
//
//	}
type RequesterReturnElidedMoq struct {
	// GetFunc mocks the Get method.
	GetFunc func(path string) (int, int, int, error)

	// PutFunc mocks the Put method.
	PutFunc func(path string) (int, error)

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
			// Path is the path argument value.
			Path string
		}
		// Put holds details about calls to the Put method.
		Put []struct {
			// Path is the path argument value.
			Path string
		}
	}
	lockGet sync.RWMutex
	lockPut sync.RWMutex
}

// Get calls GetFunc.
func (mock *RequesterReturnElidedMoq) Get(path string) (int, int, int, error) {
	if mock.GetFunc == nil {
		panic("RequesterReturnElidedMoq.GetFunc: method is nil but RequesterReturnElided.Get was just called")
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
//	len(mockedRequesterReturnElided.GetCalls())
func (mock *RequesterReturnElidedMoq) GetCalls() []struct {
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
func (mock *RequesterReturnElidedMoq) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// Put calls PutFunc.
func (mock *RequesterReturnElidedMoq) Put(path string) (int, error) {
	if mock.PutFunc == nil {
		panic("RequesterReturnElidedMoq.PutFunc: method is nil but RequesterReturnElided.Put was just called")
	}
	callInfo := struct {
		Path string
	}{
		Path: path,
	}
	mock.lockPut.Lock()
	mock.calls.Put = append(mock.calls.Put, callInfo)
	mock.lockPut.Unlock()
	return mock.PutFunc(path)
}

// PutCalls gets all the calls that were made to Put.
// Check the length with:
//
//	len(mockedRequesterReturnElided.PutCalls())
func (mock *RequesterReturnElidedMoq) PutCalls() []struct {
	Path string
} {
	var calls []struct {
		Path string
	}
	mock.lockPut.RLock()
	calls = mock.calls.Put
	mock.lockPut.RUnlock()
	return calls
}

// ResetPutCalls reset all the calls that were made to Put.
func (mock *RequesterReturnElidedMoq) ResetPutCalls() {
	mock.lockPut.Lock()
	mock.calls.Put = nil
	mock.lockPut.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *RequesterReturnElidedMoq) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()

	mock.lockPut.Lock()
	mock.calls.Put = nil
	mock.lockPut.Unlock()
}
