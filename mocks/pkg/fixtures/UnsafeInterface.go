// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	"unsafe"

	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// UnsafeInterface is an autogenerated mock type for the UnsafeInterface type
type UnsafeInterface struct {
	mock.Mock
}

// Do provides a mock function with given fields: ptr
func (_m *UnsafeInterface) Do(ptr *unsafe.Pointer) {
	_m.Called(ptr)
}

// NewUnsafeInterface creates a new instance of UnsafeInterface. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewUnsafeInterface(t testing.TB) *UnsafeInterface {
	mock := &UnsafeInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
