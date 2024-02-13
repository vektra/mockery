// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// Blank is a mock implementation of Blank.
//
//	func TestSomethingThatUsesBlank(t *testing.T) {
//
//		// make and configure a mocked Blank
//		mockedBlank := &Blank{
//			CreateFunc: func(x interface{}) error {
//				panic("mock out the Create method")
//			},
//		}
//
//		// use mockedBlank in code that requires Blank
//		// and then make assertions.
//
//	}
type Blank struct {
	// CreateFunc mocks the Create method.
	CreateFunc func(x interface{}) error

	// calls tracks calls to the methods.
	calls struct {
		// Create holds details about calls to the Create method.
		Create []struct {
			// X is the x argument value.
			X interface{}
		}
	}
	lockCreate sync.RWMutex
}

// Create calls CreateFunc.
func (mock *Blank) Create(x interface{}) error {
	if mock.CreateFunc == nil {
		panic("Blank.CreateFunc: method is nil but Blank.Create was just called")
	}
	callInfo := struct {
		X interface{}
	}{
		X: x,
	}
	mock.lockCreate.Lock()
	mock.calls.Create = append(mock.calls.Create, callInfo)
	mock.lockCreate.Unlock()
	return mock.CreateFunc(x)
}

// CreateCalls gets all the calls that were made to Create.
// Check the length with:
//
//	len(mockedBlank.CreateCalls())
func (mock *Blank) CreateCalls() []struct {
	X interface{}
} {
	var calls []struct {
		X interface{}
	}
	mock.lockCreate.RLock()
	calls = mock.calls.Create
	mock.lockCreate.RUnlock()
	return calls
}

// ResetCreateCalls reset all the calls that were made to Create.
func (mock *Blank) ResetCreateCalls() {
	mock.lockCreate.Lock()
	mock.calls.Create = nil
	mock.lockCreate.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *Blank) ResetCalls() {
	mock.lockCreate.Lock()
	mock.calls.Create = nil
	mock.lockCreate.Unlock()
}
