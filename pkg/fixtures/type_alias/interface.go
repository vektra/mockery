package type_alias

import "github.com/vektra/mockery/v2/pkg/fixtures/type_alias/subpkg"

type Type = int
type S = subpkg.S

type Interface1 interface {
	Foo() Type
}

type Interface2 interface {
	F(Type, S, subpkg.S)
}
