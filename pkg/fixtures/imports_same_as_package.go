package test

import (
	test "github.com/vektra/mockery/v3/pkg/fixtures/redefined_type_b"
)

type C int

type ImportsSameAsPackage interface {
	A() test.B
	B() KeyManager
	C(C)
}
