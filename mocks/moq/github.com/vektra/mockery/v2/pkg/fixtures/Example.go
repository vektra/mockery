// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"net/http"
	"sync"

	my_http "github.com/vektra/mockery/v2/pkg/fixtures/http"
)

// Example is a mock implementation of Example.
//
//	func TestSomethingThatUsesExample(t *testing.T) {
//
//		// make and configure a mocked Example
//		mockedExample := &Example{
//			AFunc: func() http.Flusher {
//				panic("mock out the A method")
//			},
//			BFunc: func(fixtureshttp string) my_http.MyStruct {
//				panic("mock out the B method")
//			},
//		}
//
//		// use mockedExample in code that requires Example
//		// and then make assertions.
//
//	}
type Example struct {
	// AFunc mocks the A method.
	AFunc func() http.Flusher

	// BFunc mocks the B method.
	BFunc func(fixtureshttp string) my_http.MyStruct

	// calls tracks calls to the methods.
	calls struct {
		// A holds details about calls to the A method.
		A []struct {
		}
		// B holds details about calls to the B method.
		B []struct {
			// Fixtureshttp is the fixtureshttp argument value.
			Fixtureshttp string
		}
	}
	lockA sync.RWMutex
	lockB sync.RWMutex
}

// A calls AFunc.
func (mock *Example) A() http.Flusher {
	if mock.AFunc == nil {
		panic("Example.AFunc: method is nil but Example.A was just called")
	}
	callInfo := struct {
	}{}
	mock.lockA.Lock()
	mock.calls.A = append(mock.calls.A, callInfo)
	mock.lockA.Unlock()
	return mock.AFunc()
}

// ACalls gets all the calls that were made to A.
// Check the length with:
//
//	len(mockedExample.ACalls())
func (mock *Example) ACalls() []struct {
} {
	var calls []struct {
	}
	mock.lockA.RLock()
	calls = mock.calls.A
	mock.lockA.RUnlock()
	return calls
}

// ResetACalls reset all the calls that were made to A.
func (mock *Example) ResetACalls() {
	mock.lockA.Lock()
	mock.calls.A = nil
	mock.lockA.Unlock()
}

// B calls BFunc.
func (mock *Example) B(fixtureshttp string) my_http.MyStruct {
	if mock.BFunc == nil {
		panic("Example.BFunc: method is nil but Example.B was just called")
	}
	callInfo := struct {
		Fixtureshttp string
	}{
		Fixtureshttp: fixtureshttp,
	}
	mock.lockB.Lock()
	mock.calls.B = append(mock.calls.B, callInfo)
	mock.lockB.Unlock()
	return mock.BFunc(fixtureshttp)
}

// BCalls gets all the calls that were made to B.
// Check the length with:
//
//	len(mockedExample.BCalls())
func (mock *Example) BCalls() []struct {
	Fixtureshttp string
} {
	var calls []struct {
		Fixtureshttp string
	}
	mock.lockB.RLock()
	calls = mock.calls.B
	mock.lockB.RUnlock()
	return calls
}

// ResetBCalls reset all the calls that were made to B.
func (mock *Example) ResetBCalls() {
	mock.lockB.Lock()
	mock.calls.B = nil
	mock.lockB.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *Example) ResetCalls() {
	mock.lockA.Lock()
	mock.calls.A = nil
	mock.lockA.Unlock()

	mock.lockB.Lock()
	mock.calls.B = nil
	mock.lockB.Unlock()
}
