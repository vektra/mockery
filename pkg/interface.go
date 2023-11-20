package pkg

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Interface type represents the target type that we will generate a mock for.
// It could be an interface, or a function type.
// Function type emulates: an interface it has 1 method with the function signature
// and a general name, e.g. "Execute".
type Interface struct {
	Name            string // Name of the type to be mocked.
	QualifiedName   string // Path to the package of the target type.
	FileName        string
	File            *ast.File
	PackagesPackage *packages.Package
	Pkg             TypesPackage
	NamedType       *types.Named
	IsFunction      bool             // If true, this instance represents a function, otherwise it's an interface.
	ActualInterface *types.Interface // Holds the actual interface type, in case it's an interface.
	SingleFunction  *Method          // Holds the function type information, in case it's a function type.
}

func (iface *Interface) Methods() []*Method {
	if iface.IsFunction {
		return []*Method{iface.SingleFunction}
	}
	methods := make([]*Method, iface.ActualInterface.NumMethods())
	for i := 0; i < iface.ActualInterface.NumMethods(); i++ {
		fn := iface.ActualInterface.Method(i)
		methods[i] = &Method{Name: fn.Name(), Signature: fn.Type().(*types.Signature)}
	}
	return methods
}
