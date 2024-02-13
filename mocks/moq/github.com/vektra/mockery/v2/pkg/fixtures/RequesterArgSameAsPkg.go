// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// RequesterArgSameAsPkgMock is a mock implementation of RequesterArgSameAsPkg.
//
//	func TestSomethingThatUsesRequesterArgSameAsPkg(t *testing.T) {
//
//		// make and configure a mocked RequesterArgSameAsPkg
//		mockedRequesterArgSameAsPkg := &RequesterArgSameAsPkgMock{
//			GetFunc: func(test string)  {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedRequesterArgSameAsPkg in code that requires RequesterArgSameAsPkg
//		// and then make assertions.
//
//	}
type RequesterArgSameAsPkgMock struct {
	// GetFunc mocks the Get method.
	GetFunc func(test string)

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
			// Test is the test argument value.
			Test string
		}
	}
	lockGet sync.RWMutex
}

// Get calls GetFunc.
func (mock *RequesterArgSameAsPkgMock) Get(test string) {
	if mock.GetFunc == nil {
		panic("RequesterArgSameAsPkgMock.GetFunc: method is nil but RequesterArgSameAsPkg.Get was just called")
	}
	callInfo := struct {
		Test string
	}{
		Test: test,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	mock.GetFunc(test)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//
//	len(mockedRequesterArgSameAsPkg.GetCalls())
func (mock *RequesterArgSameAsPkgMock) GetCalls() []struct {
	Test string
} {
	var calls []struct {
		Test string
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// ResetGetCalls reset all the calls that were made to Get.
func (mock *RequesterArgSameAsPkgMock) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *RequesterArgSameAsPkgMock) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}
