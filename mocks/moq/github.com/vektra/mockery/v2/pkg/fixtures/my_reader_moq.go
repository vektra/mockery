// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// Ensure, that MyReaderMoq does implement MyReader.
// If this is not the case, regenerate this file with moq.
var _ MyReader = &MyReaderMoq{}

// MyReaderMoq is a mock implementation of MyReader.
//
//	func TestSomethingThatUsesMyReader(t *testing.T) {
//
//		// make and configure a mocked MyReader
//		mockedMyReader := &MyReaderMoq{
//			ReadFunc: func(p []byte) (int, error) {
//				panic("mock out the Read method")
//			},
//		}
//
//		// use mockedMyReader in code that requires MyReader
//		// and then make assertions.
//
//	}
type MyReaderMoq struct {
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
func (mock *MyReaderMoq) Read(p []byte) (int, error) {
	if mock.ReadFunc == nil {
		panic("MyReaderMoq.ReadFunc: method is nil but MyReader.Read was just called")
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
func (mock *MyReaderMoq) ReadCalls() []struct {
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

// ResetReadCalls reset all the calls that were made to Read.
func (mock *MyReaderMoq) ResetReadCalls() {
	mock.lockRead.Lock()
	mock.calls.Read = nil
	mock.lockRead.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *MyReaderMoq) ResetCalls() {
	mock.lockRead.Lock()
	mock.calls.Read = nil
	mock.lockRead.Unlock()
}
