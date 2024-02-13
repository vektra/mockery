// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// AsyncProducerMock is a mock implementation of AsyncProducer.
//
//	func TestSomethingThatUsesAsyncProducer(t *testing.T) {
//
//		// make and configure a mocked AsyncProducer
//		mockedAsyncProducer := &AsyncProducerMock{
//			InputFunc: func() chan<- bool {
//				panic("mock out the Input method")
//			},
//			OutputFunc: func() <-chan bool {
//				panic("mock out the Output method")
//			},
//			WhateverFunc: func() chan bool {
//				panic("mock out the Whatever method")
//			},
//		}
//
//		// use mockedAsyncProducer in code that requires AsyncProducer
//		// and then make assertions.
//
//	}
type AsyncProducerMock struct {
	// InputFunc mocks the Input method.
	InputFunc func() chan<- bool

	// OutputFunc mocks the Output method.
	OutputFunc func() <-chan bool

	// WhateverFunc mocks the Whatever method.
	WhateverFunc func() chan bool

	// calls tracks calls to the methods.
	calls struct {
		// Input holds details about calls to the Input method.
		Input []struct {
		}
		// Output holds details about calls to the Output method.
		Output []struct {
		}
		// Whatever holds details about calls to the Whatever method.
		Whatever []struct {
		}
	}
	lockInput    sync.RWMutex
	lockOutput   sync.RWMutex
	lockWhatever sync.RWMutex
}

// Input calls InputFunc.
func (mock *AsyncProducerMock) Input() chan<- bool {
	if mock.InputFunc == nil {
		panic("AsyncProducerMock.InputFunc: method is nil but AsyncProducer.Input was just called")
	}
	callInfo := struct {
	}{}
	mock.lockInput.Lock()
	mock.calls.Input = append(mock.calls.Input, callInfo)
	mock.lockInput.Unlock()
	return mock.InputFunc()
}

// InputCalls gets all the calls that were made to Input.
// Check the length with:
//
//	len(mockedAsyncProducer.InputCalls())
func (mock *AsyncProducerMock) InputCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockInput.RLock()
	calls = mock.calls.Input
	mock.lockInput.RUnlock()
	return calls
}

// ResetInputCalls reset all the calls that were made to Input.
func (mock *AsyncProducerMock) ResetInputCalls() {
	mock.lockInput.Lock()
	mock.calls.Input = nil
	mock.lockInput.Unlock()
}

// Output calls OutputFunc.
func (mock *AsyncProducerMock) Output() <-chan bool {
	if mock.OutputFunc == nil {
		panic("AsyncProducerMock.OutputFunc: method is nil but AsyncProducer.Output was just called")
	}
	callInfo := struct {
	}{}
	mock.lockOutput.Lock()
	mock.calls.Output = append(mock.calls.Output, callInfo)
	mock.lockOutput.Unlock()
	return mock.OutputFunc()
}

// OutputCalls gets all the calls that were made to Output.
// Check the length with:
//
//	len(mockedAsyncProducer.OutputCalls())
func (mock *AsyncProducerMock) OutputCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockOutput.RLock()
	calls = mock.calls.Output
	mock.lockOutput.RUnlock()
	return calls
}

// ResetOutputCalls reset all the calls that were made to Output.
func (mock *AsyncProducerMock) ResetOutputCalls() {
	mock.lockOutput.Lock()
	mock.calls.Output = nil
	mock.lockOutput.Unlock()
}

// Whatever calls WhateverFunc.
func (mock *AsyncProducerMock) Whatever() chan bool {
	if mock.WhateverFunc == nil {
		panic("AsyncProducerMock.WhateverFunc: method is nil but AsyncProducer.Whatever was just called")
	}
	callInfo := struct {
	}{}
	mock.lockWhatever.Lock()
	mock.calls.Whatever = append(mock.calls.Whatever, callInfo)
	mock.lockWhatever.Unlock()
	return mock.WhateverFunc()
}

// WhateverCalls gets all the calls that were made to Whatever.
// Check the length with:
//
//	len(mockedAsyncProducer.WhateverCalls())
func (mock *AsyncProducerMock) WhateverCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockWhatever.RLock()
	calls = mock.calls.Whatever
	mock.lockWhatever.RUnlock()
	return calls
}

// ResetWhateverCalls reset all the calls that were made to Whatever.
func (mock *AsyncProducerMock) ResetWhateverCalls() {
	mock.lockWhatever.Lock()
	mock.calls.Whatever = nil
	mock.lockWhatever.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *AsyncProducerMock) ResetCalls() {
	mock.lockInput.Lock()
	mock.calls.Input = nil
	mock.lockInput.Unlock()

	mock.lockOutput.Lock()
	mock.calls.Output = nil
	mock.lockOutput.Unlock()

	mock.lockWhatever.Lock()
	mock.calls.Whatever = nil
	mock.lockWhatever.Unlock()
}
