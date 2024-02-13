// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"

	test "github.com/vektra/mockery/v2/pkg/fixtures"
)

// UsesOtherPkgIfaceMock is a mock implementation of UsesOtherPkgIface.
//
//	func TestSomethingThatUsesUsesOtherPkgIface(t *testing.T) {
//
//		// make and configure a mocked UsesOtherPkgIface
//		mockedUsesOtherPkgIface := &UsesOtherPkgIfaceMock{
//			DoSomethingElseFunc: func(obj test.Sibling)  {
//				panic("mock out the DoSomethingElse method")
//			},
//		}
//
//		// use mockedUsesOtherPkgIface in code that requires UsesOtherPkgIface
//		// and then make assertions.
//
//	}
type UsesOtherPkgIfaceMock struct {
	// DoSomethingElseFunc mocks the DoSomethingElse method.
	DoSomethingElseFunc func(obj test.Sibling)

	// calls tracks calls to the methods.
	calls struct {
		// DoSomethingElse holds details about calls to the DoSomethingElse method.
		DoSomethingElse []struct {
			// Obj is the obj argument value.
			Obj test.Sibling
		}
	}
	lockDoSomethingElse sync.RWMutex
}

// DoSomethingElse calls DoSomethingElseFunc.
func (mock *UsesOtherPkgIfaceMock) DoSomethingElse(obj test.Sibling) {
	if mock.DoSomethingElseFunc == nil {
		panic("UsesOtherPkgIfaceMock.DoSomethingElseFunc: method is nil but UsesOtherPkgIface.DoSomethingElse was just called")
	}
	callInfo := struct {
		Obj test.Sibling
	}{
		Obj: obj,
	}
	mock.lockDoSomethingElse.Lock()
	mock.calls.DoSomethingElse = append(mock.calls.DoSomethingElse, callInfo)
	mock.lockDoSomethingElse.Unlock()
	mock.DoSomethingElseFunc(obj)
}

// DoSomethingElseCalls gets all the calls that were made to DoSomethingElse.
// Check the length with:
//
//	len(mockedUsesOtherPkgIface.DoSomethingElseCalls())
func (mock *UsesOtherPkgIfaceMock) DoSomethingElseCalls() []struct {
	Obj test.Sibling
} {
	var calls []struct {
		Obj test.Sibling
	}
	mock.lockDoSomethingElse.RLock()
	calls = mock.calls.DoSomethingElse
	mock.lockDoSomethingElse.RUnlock()
	return calls
}

// ResetDoSomethingElseCalls reset all the calls that were made to DoSomethingElse.
func (mock *UsesOtherPkgIfaceMock) ResetDoSomethingElseCalls() {
	mock.lockDoSomethingElse.Lock()
	mock.calls.DoSomethingElse = nil
	mock.lockDoSomethingElse.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *UsesOtherPkgIfaceMock) ResetCalls() {
	mock.lockDoSomethingElse.Lock()
	mock.calls.DoSomethingElse = nil
	mock.lockDoSomethingElse.Unlock()
}
