// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	"sync"
)

// Ensure, that FooerMoq does implement Fooer.
// If this is not the case, regenerate this file with moq.
var _ Fooer = &FooerMoq{}

// FooerMoq is a mock implementation of Fooer.
//
//	func TestSomethingThatUsesFooer(t *testing.T) {
//
//		// make and configure a mocked Fooer
//		mockedFooer := &FooerMoq{
//			BarFunc: func(f func([]int))  {
//				panic("mock out the Bar method")
//			},
//			BazFunc: func(path string) func(x string) string {
//				panic("mock out the Baz method")
//			},
//			FooFunc: func(f func(x string) string) error {
//				panic("mock out the Foo method")
//			},
//		}
//
//		// use mockedFooer in code that requires Fooer
//		// and then make assertions.
//
//	}
type FooerMoq struct {
	// BarFunc mocks the Bar method.
	BarFunc func(f func([]int))

	// BazFunc mocks the Baz method.
	BazFunc func(path string) func(x string) string

	// FooFunc mocks the Foo method.
	FooFunc func(f func(x string) string) error

	// calls tracks calls to the methods.
	calls struct {
		// Bar holds details about calls to the Bar method.
		Bar []struct {
			// F is the f argument value.
			F func([]int)
		}
		// Baz holds details about calls to the Baz method.
		Baz []struct {
			// Path is the path argument value.
			Path string
		}
		// Foo holds details about calls to the Foo method.
		Foo []struct {
			// F is the f argument value.
			F func(x string) string
		}
	}
	lockBar sync.RWMutex
	lockBaz sync.RWMutex
	lockFoo sync.RWMutex
}

// Bar calls BarFunc.
func (mock *FooerMoq) Bar(f func([]int)) {
	if mock.BarFunc == nil {
		panic("FooerMoq.BarFunc: method is nil but Fooer.Bar was just called")
	}
	callInfo := struct {
		F func([]int)
	}{
		F: f,
	}
	mock.lockBar.Lock()
	mock.calls.Bar = append(mock.calls.Bar, callInfo)
	mock.lockBar.Unlock()
	mock.BarFunc(f)
}

// BarCalls gets all the calls that were made to Bar.
// Check the length with:
//
//	len(mockedFooer.BarCalls())
func (mock *FooerMoq) BarCalls() []struct {
	F func([]int)
} {
	var calls []struct {
		F func([]int)
	}
	mock.lockBar.RLock()
	calls = mock.calls.Bar
	mock.lockBar.RUnlock()
	return calls
}

// ResetBarCalls reset all the calls that were made to Bar.
func (mock *FooerMoq) ResetBarCalls() {
	mock.lockBar.Lock()
	mock.calls.Bar = nil
	mock.lockBar.Unlock()
}

// Baz calls BazFunc.
func (mock *FooerMoq) Baz(path string) func(x string) string {
	if mock.BazFunc == nil {
		panic("FooerMoq.BazFunc: method is nil but Fooer.Baz was just called")
	}
	callInfo := struct {
		Path string
	}{
		Path: path,
	}
	mock.lockBaz.Lock()
	mock.calls.Baz = append(mock.calls.Baz, callInfo)
	mock.lockBaz.Unlock()
	return mock.BazFunc(path)
}

// BazCalls gets all the calls that were made to Baz.
// Check the length with:
//
//	len(mockedFooer.BazCalls())
func (mock *FooerMoq) BazCalls() []struct {
	Path string
} {
	var calls []struct {
		Path string
	}
	mock.lockBaz.RLock()
	calls = mock.calls.Baz
	mock.lockBaz.RUnlock()
	return calls
}

// ResetBazCalls reset all the calls that were made to Baz.
func (mock *FooerMoq) ResetBazCalls() {
	mock.lockBaz.Lock()
	mock.calls.Baz = nil
	mock.lockBaz.Unlock()
}

// Foo calls FooFunc.
func (mock *FooerMoq) Foo(f func(x string) string) error {
	if mock.FooFunc == nil {
		panic("FooerMoq.FooFunc: method is nil but Fooer.Foo was just called")
	}
	callInfo := struct {
		F func(x string) string
	}{
		F: f,
	}
	mock.lockFoo.Lock()
	mock.calls.Foo = append(mock.calls.Foo, callInfo)
	mock.lockFoo.Unlock()
	return mock.FooFunc(f)
}

// FooCalls gets all the calls that were made to Foo.
// Check the length with:
//
//	len(mockedFooer.FooCalls())
func (mock *FooerMoq) FooCalls() []struct {
	F func(x string) string
} {
	var calls []struct {
		F func(x string) string
	}
	mock.lockFoo.RLock()
	calls = mock.calls.Foo
	mock.lockFoo.RUnlock()
	return calls
}

// ResetFooCalls reset all the calls that were made to Foo.
func (mock *FooerMoq) ResetFooCalls() {
	mock.lockFoo.Lock()
	mock.calls.Foo = nil
	mock.lockFoo.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *FooerMoq) ResetCalls() {
	mock.lockBar.Lock()
	mock.calls.Bar = nil
	mock.lockBar.Unlock()

	mock.lockBaz.Lock()
	mock.calls.Baz = nil
	mock.lockBaz.Unlock()

	mock.lockFoo.Lock()
	mock.calls.Foo = nil
	mock.lockFoo.Unlock()
}
