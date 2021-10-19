package test

import (
	"github.com/pendo-io/b2h-mockgen/pkg/fixtures/test"
)

type C int

type ImportsSameAsPackage interface {
	A() test.B
	B() KeyManager
	C(C)
}
