// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package test

import (
	fixtures "github.com/vektra/mockery/v2/pkg/fixtures"
	redefinedtypeb "github.com/vektra/mockery/v2/pkg/fixtures/redefined_type_b"
	"sync"
)

// ImportsSameAsPackageMock is a mock implementation of test.ImportsSameAsPackage.
//
//	func TestSomethingThatUsesImportsSameAsPackage(t *testing.T) {
//
//		// make and configure a mocked test.ImportsSameAsPackage
//		mockedImportsSameAsPackage := &ImportsSameAsPackageMock{
//			AFunc: func() redefinedtypeb.B {
//				panic("mock out the A method")
//			},
//			BFunc: func() fixtures.KeyManager {
//				panic("mock out the B method")
//			},
//			CFunc: func(c fixtures.C)  {
//				panic("mock out the C method")
//			},
//		}
//
//		// use mockedImportsSameAsPackage in code that requires test.ImportsSameAsPackage
//		// and then make assertions.
//
//	}
type ImportsSameAsPackageMock struct {
	// AFunc mocks the A method.
	AFunc func() redefinedtypeb.B

	// BFunc mocks the B method.
	BFunc func() fixtures.KeyManager

	// CFunc mocks the C method.
	CFunc func(c fixtures.C)

	// calls tracks calls to the methods.
	calls struct {
		// A holds details about calls to the A method.
		A []struct {
		}
		// B holds details about calls to the B method.
		B []struct {
		}
		// C holds details about calls to the C method.
		C []struct {
			// C is the c argument value.
			C fixtures.C
		}
	}
	lockA sync.RWMutex
	lockB sync.RWMutex
	lockC sync.RWMutex
}

// A calls AFunc.
func (mock *ImportsSameAsPackageMock) A() redefinedtypeb.B {
	if mock.AFunc == nil {
		panic("ImportsSameAsPackageMock.AFunc: method is nil but ImportsSameAsPackage.A was just called")
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
//	len(mockedImportsSameAsPackage.ACalls())
func (mock *ImportsSameAsPackageMock) ACalls() []struct {
} {
	var calls []struct {
	}
	mock.lockA.RLock()
	calls = mock.calls.A
	mock.lockA.RUnlock()
	return calls
}

// ResetACalls reset all the calls that were made to A.
func (mock *ImportsSameAsPackageMock) ResetACalls() {
	mock.lockA.Lock()
	mock.calls.A = nil
	mock.lockA.Unlock()
}

// B calls BFunc.
func (mock *ImportsSameAsPackageMock) B() fixtures.KeyManager {
	if mock.BFunc == nil {
		panic("ImportsSameAsPackageMock.BFunc: method is nil but ImportsSameAsPackage.B was just called")
	}
	callInfo := struct {
	}{}
	mock.lockB.Lock()
	mock.calls.B = append(mock.calls.B, callInfo)
	mock.lockB.Unlock()
	return mock.BFunc()
}

// BCalls gets all the calls that were made to B.
// Check the length with:
//
//	len(mockedImportsSameAsPackage.BCalls())
func (mock *ImportsSameAsPackageMock) BCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockB.RLock()
	calls = mock.calls.B
	mock.lockB.RUnlock()
	return calls
}

// ResetBCalls reset all the calls that were made to B.
func (mock *ImportsSameAsPackageMock) ResetBCalls() {
	mock.lockB.Lock()
	mock.calls.B = nil
	mock.lockB.Unlock()
}

// C calls CFunc.
func (mock *ImportsSameAsPackageMock) C(c fixtures.C) {
	if mock.CFunc == nil {
		panic("ImportsSameAsPackageMock.CFunc: method is nil but ImportsSameAsPackage.C was just called")
	}
	callInfo := struct {
		C fixtures.C
	}{
		C: c,
	}
	mock.lockC.Lock()
	mock.calls.C = append(mock.calls.C, callInfo)
	mock.lockC.Unlock()
	mock.CFunc(c)
}

// CCalls gets all the calls that were made to C.
// Check the length with:
//
//	len(mockedImportsSameAsPackage.CCalls())
func (mock *ImportsSameAsPackageMock) CCalls() []struct {
	C fixtures.C
} {
	var calls []struct {
		C fixtures.C
	}
	mock.lockC.RLock()
	calls = mock.calls.C
	mock.lockC.RUnlock()
	return calls
}

// ResetCCalls reset all the calls that were made to C.
func (mock *ImportsSameAsPackageMock) ResetCCalls() {
	mock.lockC.Lock()
	mock.calls.C = nil
	mock.lockC.Unlock()
}

// ResetCalls reset all the calls that were made to all mocked methods.
func (mock *ImportsSameAsPackageMock) ResetCalls() {
	mock.lockA.Lock()
	mock.calls.A = nil
	mock.lockA.Unlock()

	mock.lockB.Lock()
	mock.calls.B = nil
	mock.lockB.Unlock()

	mock.lockC.Lock()
	mock.calls.C = nil
	mock.lockC.Unlock()
}
