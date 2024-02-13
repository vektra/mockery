// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"encoding/json"
	"sync"
)

// RequesterArgSameAsNamedImport is a mock implementation of RequesterArgSameAsNamedImport.
//
//	func TestSomethingThatUsesRequesterArgSameAsNamedImport(t *testing.T) {
//
//		// make and configure a mocked RequesterArgSameAsNamedImport
//		mockedRequesterArgSameAsNamedImport := &RequesterArgSameAsNamedImport{
//			GetFunc: func(jsonMoqParam string) *json.RawMessage {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedRequesterArgSameAsNamedImport in code that requires RequesterArgSameAsNamedImport
//		// and then make assertions.
//
//	}
type RequesterArgSameAsNamedImport struct {
	// GetFunc mocks the Get method.
	GetFunc func(jsonMoqParam string) *json.RawMessage

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
			// JsonMoqParam is the jsonMoqParam argument value.
			JsonMoqParam string
		}
	}
	lockGet sync.RWMutex
}

// Get calls GetFunc.
func (mock *RequesterArgSameAsNamedImport) Get(jsonMoqParam string) *json.RawMessage {
	if mock.GetFunc == nil {
		panic("RequesterArgSameAsNamedImport.GetFunc: method is nil but RequesterArgSameAsNamedImport.Get was just called")
	}
	callInfo := struct {
		JsonMoqParam string
	}{
		JsonMoqParam: jsonMoqParam,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(jsonMoqParam)
}

// GetCalls gets all the calls that were made to Get.
// Check the length with:
//
//	len(mockedRequesterArgSameAsNamedImport.GetCalls())
func (mock *RequesterArgSameAsNamedImport) GetCalls() []struct {
	JsonMoqParam string
} {
	var calls []struct {
		JsonMoqParam string
	}
	mock.lockGet.RLock()
	calls = mock.calls.Get
	mock.lockGet.RUnlock()
	return calls
}

// ResetGetCalls reset all the calls that were made to Get.
func (mock *RequesterArgSameAsNamedImport) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *RequesterArgSameAsNamedImport) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}
