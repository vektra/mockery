// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package testfoo

import (
	"sync"

	test "github.com/vektra/mockery/v2/pkg/fixtures"
)

// Ensure, that MyReader does implement test.MyReader.
// If this is not the case, regenerate this file with moq.
var _ test.MyReader = &MyReader{}

// MyReader is a mock implementation of test.MyReader.
//
//	func TestSomethingThatUsesMyReader(t *testing.T) {
//
//		// make and configure a mocked test.MyReader
//		mockedMyReader := &MyReader{
//			ReadFunc: func(p []byte) (int, error) {
//				panic("mock out the Read method")
//			},
//		}
//
//		// use mockedMyReader in code that requires test.MyReader
//		// and then make assertions.
//
//	}
type MyReader struct {
	// ReadFunc mocks the Read method.
	ReadFunc func(p []byte) (int, error)

	// calls tracks calls to the methods.
	calls struct {
		// Read holds details about calls to the Read method.
		Read []struct {
			// P is the p argument value.
			P []byte
		}
	}
	lockRead sync.RWMutex
}

// Read calls ReadFunc.
func (mock *MyReader) Read(p []byte) (int, error) {
	if mock.ReadFunc == nil {
		panic("MyReader.ReadFunc: method is nil but MyReader.Read was just called")
	}
	callInfo := struct {
		P []byte
	}{
		P: p,
	}
	mock.lockRead.Lock()
	mock.calls.Read = append(mock.calls.Read, callInfo)
	mock.lockRead.Unlock()
	return mock.ReadFunc(p)
}

// ReadCalls gets all the calls that were made to Read.
// Check the length with:
//
//	len(mockedMyReader.ReadCalls())
func (mock *MyReader) ReadCalls() []struct {
	P []byte
} {
	var calls []struct {
		P []byte
	}
	mock.lockRead.RLock()
	calls = mock.calls.Read
	mock.lockRead.RUnlock()
	return calls
}
