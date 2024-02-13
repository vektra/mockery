// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"net/http"
	"sync"
)

// RequesterNS is a mock implementation of RequesterNS.
//
//	func TestSomethingThatUsesRequesterNS(t *testing.T) {
//
//		// make and configure a mocked RequesterNS
//		mockedRequesterNS := &RequesterNS{
//			GetFunc: func(path string) (http.Response, error) {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedRequesterNS in code that requires RequesterNS
//		// and then make assertions.
//
//	}
type RequesterNS struct {
	// GetFunc mocks the Get method.
	GetFunc func(path string) (http.Response, error)

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
func (mock *RequesterNS) Get(path string) (http.Response, error) {
	if mock.GetFunc == nil {
		panic("RequesterNS.GetFunc: method is nil but RequesterNS.Get was just called")
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
//	len(mockedRequesterNS.GetCalls())
func (mock *RequesterNS) GetCalls() []struct {
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
func (mock *RequesterNS) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *RequesterNS) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}
