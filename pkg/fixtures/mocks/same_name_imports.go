// Code generated by B2H-MockGen v0.0.0-dev. EDIT AT YOUR OWN PERIL.

package mocks

import (
	http "net/http"

	fixtureshttp "github.com/pendo-io/b2h-mockgen/pkg/fixtures/http"

	mock "github.com/stretchr/testify/mock"
)

type ExampleMock struct {
	mock.Mock
}

func (_m *ExampleMock) A() http.Flusher {
	args := _m.Called()
	return args.Get(0).(http.Flusher)
}
func (_m *ExampleMock) B(_a0 string) fixtureshttp.MyStruct {
	args := _m.Called(_a0)
	return args.Get(0).(fixtureshttp.MyStruct)
}
