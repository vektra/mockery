// Code generated by B2H-MockGen v0.0.0-dev. EDIT AT YOUR OWN PERIL.

package b2hmock

import (
	mock "github.com/stretchr/testify/mock"
	test "github.com/vektra/mockery/v2/pkg/fixtures"
)

type UsesOtherPkgIfaceMock struct {
	mock.Mock
}

func (_m *UsesOtherPkgIfaceMock) DoSomethingElse(obj test.Sibling) {
	_m.Called(obj)
}