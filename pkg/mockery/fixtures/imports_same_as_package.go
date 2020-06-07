package test

import (
	"github.com/vektra/mockery/pkg/mockery/fixtures/test"
)

type C int

type ImportsSameAsPackage interface {
	A() test.B
	B() KeyManager
	C(C)
}
