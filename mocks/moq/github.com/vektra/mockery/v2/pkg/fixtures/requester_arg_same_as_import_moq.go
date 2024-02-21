// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"encoding/json"
	"sync"
)

// Ensure, that RequesterArgSameAsImportMoq does implement RequesterArgSameAsImport.
// If this is not the case, regenerate this file with moq.
var _ RequesterArgSameAsImport = &RequesterArgSameAsImportMoq{}

// RequesterArgSameAsImportMoq is a mock implementation of RequesterArgSameAsImport.
//
//	func TestSomethingThatUsesRequesterArgSameAsImport(t *testing.T) {
//
//		// make and configure a mocked RequesterArgSameAsImport
//		mockedRequesterArgSameAsImport := &RequesterArgSameAsImportMoq{
//			GetFunc: func(jsonMoqParam string) *json.RawMessage {
//				panic("mock out the Get method")
//			},
//		}
//
//		// use mockedRequesterArgSameAsImport in code that requires RequesterArgSameAsImport
//		// and then make assertions.
//
//	}
type RequesterArgSameAsImportMoq struct {
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
func (mock *RequesterArgSameAsImportMoq) Get(jsonMoqParam string) *json.RawMessage {
	if mock.GetFunc == nil {
		panic("RequesterArgSameAsImportMoq.GetFunc: method is nil but RequesterArgSameAsImport.Get was just called")
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
//	len(mockedRequesterArgSameAsImport.GetCalls())
func (mock *RequesterArgSameAsImportMoq) GetCalls() []struct {
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
func (mock *RequesterArgSameAsImportMoq) ResetGetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *RequesterArgSameAsImportMoq) ResetCalls() {
	mock.lockGet.Lock()
	mock.calls.Get = nil
	mock.lockGet.Unlock()
}
