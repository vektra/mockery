// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"

	test "github.com/vektra/mockery/v2/pkg/fixtures"
)

// Ensure, that AMoq does implement A.
// If this is not the case, regenerate this file with moq.
var _ A = &AMoq{}

// AMoq is a mock implementation of A.
//
//	func TestSomethingThatUsesA(t *testing.T) {
//
//		// make and configure a mocked A
//		mockedA := &AMoq{
//			CallFunc: func() (test.B, error) {
//				panic("mock out the Call method")
//			},
//		}
//
//		// use mockedA in code that requires A
//		// and then make assertions.
//
//	}
type AMoq struct {
	// CallFunc mocks the Call method.
	CallFunc func() (test.B, error)

	// calls tracks calls to the methods.
	calls struct {
		// Call holds details about calls to the Call method.
		Call []struct {
		}
	}
	lockCall sync.RWMutex
}

// Call calls CallFunc.
func (mock *AMoq) Call() (test.B, error) {
	if mock.CallFunc == nil {
		panic("AMoq.CallFunc: method is nil but A.Call was just called")
	}
	callInfo := struct {
	}{}
	mock.lockCall.Lock()
	mock.calls.Call = append(mock.calls.Call, callInfo)
	mock.lockCall.Unlock()
	return mock.CallFunc()
}

// CallCalls gets all the calls that were made to Call.
// Check the length with:
//
//	len(mockedA.CallCalls())
func (mock *AMoq) CallCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockCall.RLock()
	calls = mock.calls.Call
	mock.lockCall.RUnlock()
	return calls
}

// ResetCallCalls reset all the calls that were made to Call.
func (mock *AMoq) ResetCallCalls() {
	mock.lockCall.Lock()
	mock.calls.Call = nil
	mock.lockCall.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *AMoq) ResetCalls() {
	mock.lockCall.Lock()
	mock.calls.Call = nil
	mock.lockCall.Unlock()
}
