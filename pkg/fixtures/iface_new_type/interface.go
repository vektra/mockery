package iface_new_type

import "github.com/vektra/mockery/v3/pkg/fixtures/iface_new_type/subpkg"

type Interface1 interface {
	Method1()
}

type (
	Interface2 Interface1
	Interface3 subpkg.SubPkgInterface
)
