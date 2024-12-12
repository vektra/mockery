// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// ExpecterMock is a mock implementation of test.Expecter.
//
//	func TestSomethingThatUsesExpecter(t *testing.T) {
//
//		// make and configure a mocked test.Expecter
//		mockedExpecter := &ExpecterMock{
//			ManyArgsReturnsFunc: func(str string, i int) ([]string, error) {
//				panic("mock out the ManyArgsReturns method")
//			},
//			NoArgFunc: func() string {
//				panic("mock out the NoArg method")
//			},
//			NoReturnFunc: func(str string)  {
//				panic("mock out the NoReturn method")
//			},
//			VariadicFunc: func(ints ...int) error {
//				panic("mock out the Variadic method")
//			},
//			VariadicManyFunc: func(i int, a string, intfs ...interface{}) error {
//				panic("mock out the VariadicMany method")
//			},
//		}
//
//		// use mockedExpecter in code that requires test.Expecter
//		// and then make assertions.
//
//	}
type ExpecterMock struct {
	// ManyArgsReturnsFunc mocks the ManyArgsReturns method.
	ManyArgsReturnsFunc func(str string, i int) ([]string, error)

	// NoArgFunc mocks the NoArg method.
	NoArgFunc func() string

	// NoReturnFunc mocks the NoReturn method.
	NoReturnFunc func(str string)

	// VariadicFunc mocks the Variadic method.
	VariadicFunc func(ints ...int) error

	// VariadicManyFunc mocks the VariadicMany method.
	VariadicManyFunc func(i int, a string, intfs ...interface{}) error

	// calls tracks calls to the methods.
	calls struct {
		// ManyArgsReturns holds details about calls to the ManyArgsReturns method.
		ManyArgsReturns []struct {
			// Str is the str argument value.
			Str string
			// I is the i argument value.
			I int
		}
		// NoArg holds details about calls to the NoArg method.
		NoArg []struct {
		}
		// NoReturn holds details about calls to the NoReturn method.
		NoReturn []struct {
			// Str is the str argument value.
			Str string
		}
		// Variadic holds details about calls to the Variadic method.
		Variadic []struct {
			// Ints is the ints argument value.
			Ints []int
		}
		// VariadicMany holds details about calls to the VariadicMany method.
		VariadicMany []struct {
			// I is the i argument value.
			I int
			// A is the a argument value.
			A string
			// Intfs is the intfs argument value.
			Intfs []interface{}
		}
	}
	lockManyArgsReturns sync.RWMutex
	lockNoArg           sync.RWMutex
	lockNoReturn        sync.RWMutex
	lockVariadic        sync.RWMutex
	lockVariadicMany    sync.RWMutex
}

// ManyArgsReturns calls ManyArgsReturnsFunc.
func (mock *ExpecterMock) ManyArgsReturns(str string, i int) ([]string, error) {
	if mock.ManyArgsReturnsFunc == nil {
		panic("ExpecterMock.ManyArgsReturnsFunc: method is nil but Expecter.ManyArgsReturns was just called")
	}
	callInfo := struct {
		Str string
		I   int
	}{
		Str: str,
		I:   i,
	}
	mock.lockManyArgsReturns.Lock()
	mock.calls.ManyArgsReturns = append(mock.calls.ManyArgsReturns, callInfo)
	mock.lockManyArgsReturns.Unlock()
	return mock.ManyArgsReturnsFunc(str, i)
}

// ManyArgsReturnsCalls gets all the calls that were made to ManyArgsReturns.
// Check the length with:
//
//	len(mockedExpecter.ManyArgsReturnsCalls())
func (mock *ExpecterMock) ManyArgsReturnsCalls() []struct {
	Str string
	I   int
} {
	var calls []struct {
		Str string
		I   int
	}
	mock.lockManyArgsReturns.RLock()
	calls = mock.calls.ManyArgsReturns
	mock.lockManyArgsReturns.RUnlock()
	return calls
}

// ResetManyArgsReturnsCalls reset all the calls that were made to ManyArgsReturns.
func (mock *ExpecterMock) ResetManyArgsReturnsCalls() {
	mock.lockManyArgsReturns.Lock()
	mock.calls.ManyArgsReturns = nil
	mock.lockManyArgsReturns.Unlock()
}

// NoArg calls NoArgFunc.
func (mock *ExpecterMock) NoArg() string {
	if mock.NoArgFunc == nil {
		panic("ExpecterMock.NoArgFunc: method is nil but Expecter.NoArg was just called")
	}
	callInfo := struct {
	}{}
	mock.lockNoArg.Lock()
	mock.calls.NoArg = append(mock.calls.NoArg, callInfo)
	mock.lockNoArg.Unlock()
	return mock.NoArgFunc()
}

// NoArgCalls gets all the calls that were made to NoArg.
// Check the length with:
//
//	len(mockedExpecter.NoArgCalls())
func (mock *ExpecterMock) NoArgCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockNoArg.RLock()
	calls = mock.calls.NoArg
	mock.lockNoArg.RUnlock()
	return calls
}

// ResetNoArgCalls reset all the calls that were made to NoArg.
func (mock *ExpecterMock) ResetNoArgCalls() {
	mock.lockNoArg.Lock()
	mock.calls.NoArg = nil
	mock.lockNoArg.Unlock()
}

// NoReturn calls NoReturnFunc.
func (mock *ExpecterMock) NoReturn(str string) {
	if mock.NoReturnFunc == nil {
		panic("ExpecterMock.NoReturnFunc: method is nil but Expecter.NoReturn was just called")
	}
	callInfo := struct {
		Str string
	}{
		Str: str,
	}
	mock.lockNoReturn.Lock()
	mock.calls.NoReturn = append(mock.calls.NoReturn, callInfo)
	mock.lockNoReturn.Unlock()
	mock.NoReturnFunc(str)
}

// NoReturnCalls gets all the calls that were made to NoReturn.
// Check the length with:
//
//	len(mockedExpecter.NoReturnCalls())
func (mock *ExpecterMock) NoReturnCalls() []struct {
	Str string
} {
	var calls []struct {
		Str string
	}
	mock.lockNoReturn.RLock()
	calls = mock.calls.NoReturn
	mock.lockNoReturn.RUnlock()
	return calls
}

// ResetNoReturnCalls reset all the calls that were made to NoReturn.
func (mock *ExpecterMock) ResetNoReturnCalls() {
	mock.lockNoReturn.Lock()
	mock.calls.NoReturn = nil
	mock.lockNoReturn.Unlock()
}

// Variadic calls VariadicFunc.
func (mock *ExpecterMock) Variadic(ints ...int) error {
	if mock.VariadicFunc == nil {
		panic("ExpecterMock.VariadicFunc: method is nil but Expecter.Variadic was just called")
	}
	callInfo := struct {
		Ints []int
	}{
		Ints: ints,
	}
	mock.lockVariadic.Lock()
	mock.calls.Variadic = append(mock.calls.Variadic, callInfo)
	mock.lockVariadic.Unlock()
	return mock.VariadicFunc(ints...)
}

// VariadicCalls gets all the calls that were made to Variadic.
// Check the length with:
//
//	len(mockedExpecter.VariadicCalls())
func (mock *ExpecterMock) VariadicCalls() []struct {
	Ints []int
} {
	var calls []struct {
		Ints []int
	}
	mock.lockVariadic.RLock()
	calls = mock.calls.Variadic
	mock.lockVariadic.RUnlock()
	return calls
}

// ResetVariadicCalls reset all the calls that were made to Variadic.
func (mock *ExpecterMock) ResetVariadicCalls() {
	mock.lockVariadic.Lock()
	mock.calls.Variadic = nil
	mock.lockVariadic.Unlock()
}

// VariadicMany calls VariadicManyFunc.
func (mock *ExpecterMock) VariadicMany(i int, a string, intfs ...interface{}) error {
	if mock.VariadicManyFunc == nil {
		panic("ExpecterMock.VariadicManyFunc: method is nil but Expecter.VariadicMany was just called")
	}
	callInfo := struct {
		I     int
		A     string
		Intfs []interface{}
	}{
		I:     i,
		A:     a,
		Intfs: intfs,
	}
	mock.lockVariadicMany.Lock()
	mock.calls.VariadicMany = append(mock.calls.VariadicMany, callInfo)
	mock.lockVariadicMany.Unlock()
	return mock.VariadicManyFunc(i, a, intfs...)
}

// VariadicManyCalls gets all the calls that were made to VariadicMany.
// Check the length with:
//
//	len(mockedExpecter.VariadicManyCalls())
func (mock *ExpecterMock) VariadicManyCalls() []struct {
	I     int
	A     string
	Intfs []interface{}
} {
	var calls []struct {
		I     int
		A     string
		Intfs []interface{}
	}
	mock.lockVariadicMany.RLock()
	calls = mock.calls.VariadicMany
	mock.lockVariadicMany.RUnlock()
	return calls
}

// ResetVariadicManyCalls reset all the calls that were made to VariadicMany.
func (mock *ExpecterMock) ResetVariadicManyCalls() {
	mock.lockVariadicMany.Lock()
	mock.calls.VariadicMany = nil
	mock.lockVariadicMany.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *ExpecterMock) ResetCalls() {
	mock.lockManyArgsReturns.Lock()
	mock.calls.ManyArgsReturns = nil
	mock.lockManyArgsReturns.Unlock()

	mock.lockNoArg.Lock()
	mock.calls.NoArg = nil
	mock.lockNoArg.Unlock()

	mock.lockNoReturn.Lock()
	mock.calls.NoReturn = nil
	mock.lockNoReturn.Unlock()

	mock.lockVariadic.Lock()
	mock.calls.Variadic = nil
	mock.lockVariadic.Unlock()

	mock.lockVariadicMany.Lock()
	mock.calls.VariadicMany = nil
	mock.lockVariadicMany.Unlock()
}
