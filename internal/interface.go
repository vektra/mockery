package internal

import (
	"go/ast"
	"go/token"

	"github.com/vektra/mockery/v3/config"
	"golang.org/x/tools/go/packages"
)

type Interface struct {
	Name       string // Name of the type to be mocked.
	TypeSpec   *ast.TypeSpec
	GenDecl    *ast.GenDecl
	FilePath   string
	FileSyntax *ast.File
	Pkg        *packages.Package
	Config     *config.Config
}

func NewInterface(
	name string,
	typeSpec *ast.TypeSpec,
	genDecl *ast.GenDecl,
	filepath string,
	fileSyntax *ast.File,
	pkg *packages.Package,
	config *config.Config,
) *Interface {
	return &Interface{
		Name:       name,
		TypeSpec:   typeSpec,
		GenDecl:    genDecl,
		FilePath:   filepath,
		FileSyntax: fileSyntax,
		Pkg:        pkg,
		Config:     config,
	}
}

func isBuiltInType(name string) bool {
	switch name {
	case
		"any", "interface", "bool", "byte", "complex64", "complex128",
		"error", "float32", "float64", "int", "int8", "int16", "int32", "int64",
		"rune", "string", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return true
	default:
		return false
	}
}

func isPrivateType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.SelectorExpr:
		return false
	case *ast.Ident:
		if isBuiltInType(t.Name) {
			return false
		}
		return !token.IsExported(t.Name)
	case *ast.StarExpr:
		return isPrivateType(t.X)
	case *ast.Ellipsis:
		return isPrivateType(t.Elt)
	case *ast.ChanType:
		return isPrivateType(t.Value)
	case *ast.MapType:
		return isPrivateType(t.Key) || isPrivateType(t.Value)
	case *ast.ArrayType:
		return isPrivateType(t.Elt)
	default:
		return false
	}
}

func containsPrivateTypesInInterface(ts ast.Expr) bool {
	iface, ok := ts.(*ast.InterfaceType)
	if !ok {
		return false
	}

	for _, field := range iface.Methods.List {
		if len(field.Names) == 0 {
			if private := containsPrivateTypesInInterface(field.Type); private {
				return true
			}
		}

		for _, name := range field.Names {
			if !token.IsExported(name.Name) {
				return true
			}
		}

		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		if funcType.Params != nil {
			for _, p := range funcType.Params.List {
				if isPrivateType(p.Type) {
					return true
				}
			}
		}
		if funcType.Results != nil {
			for _, r := range funcType.Results.List {
				if isPrivateType(r.Type) {
					return true
				}
			}
		}
	}

	return false
}

func (i *Interface) ContainsUnexportedTypes() bool {
	return containsPrivateTypesInInterface(i.TypeSpec.Type)
}

func (i *Interface) IsExported() bool {
	return token.IsExported(i.Name)
}
