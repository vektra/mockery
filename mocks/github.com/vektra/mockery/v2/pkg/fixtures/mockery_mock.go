// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	"github.com/vektra/mockery/v2/pkg/fixtures"
)

// NewMockA creates a new instance of MockA. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockA(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockA {
	mock := &MockA{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
