package test

import (
	"github.com/vektra/mockery/v3/pkg/fixtures/test"
)

type C int

type ImportsSameAsPackage interface {
	A() test.B
	B() KeyManager
	C(C)
}
