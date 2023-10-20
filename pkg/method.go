package pkg

import (
	"go/types"
	"path"
	"strings"
)

type Method struct {
	Name      string
	Signature *types.Signature
}

type Imports map[string]*types.Package

func (m Method) populateImports(imports Imports) {
	for i := 0; i < m.Signature.Params().Len(); i++ {
		m.importsHelper(m.Signature.Params().At(i).Type(), imports)
	}
}

// stripVendorPath strips the vendor dir prefix from a package path.
// For example we might encounter an absolute path like
// github.com/foo/bar/vendor/github.com/pkg/errors which is resolved
// to github.com/pkg/errors.
func stripVendorPath(p string) string {
	parts := strings.Split(p, "/vendor/")
	if len(parts) == 1 {
		return p
	}
	return strings.TrimLeft(path.Join(parts[1:]...), "/")
}

// importsHelper extracts all the package imports for a given type
// recursively. The imported packages by a single type can be more than
// one (ex: map[a.Type]b.Type).
func (m Method) importsHelper(elem types.Type, imports map[string]*types.Package) {
	switch t := elem.(type) {
	case *types.Named:
		if pkg := t.Obj().Pkg(); pkg != nil {
			imports[stripVendorPath(pkg.Path())] = pkg
		}
		// The imports of a Type with a TypeList must be added to the imports list
		// For example: Foo[otherpackage.Bar] , must have otherpackage imported
		if targs := t.TypeArgs(); targs != nil {
			for i := 0; i < targs.Len(); i++ {
				m.importsHelper(targs.At(i), imports)
			}
		}

	case *types.Array:
		m.importsHelper(t.Elem(), imports)

	case *types.Slice:
		m.importsHelper(t.Elem(), imports)

	case *types.Signature:
		for i := 0; i < t.Params().Len(); i++ {
			m.importsHelper(t.Params().At(i).Type(), imports)
		}
		for i := 0; i < t.Results().Len(); i++ {
			m.importsHelper(t.Results().At(i).Type(), imports)
		}

	case *types.Map:
		m.importsHelper(t.Key(), imports)
		m.importsHelper(t.Elem(), imports)

	case *types.Chan:
		m.importsHelper(t.Elem(), imports)

	case *types.Pointer:
		m.importsHelper(t.Elem(), imports)

	case *types.Struct: // anonymous struct
		for i := 0; i < t.NumFields(); i++ {
			m.importsHelper(t.Field(i).Type(), imports)
		}

	case *types.Interface: // anonymous interface
		for i := 0; i < t.NumExplicitMethods(); i++ {
			m.importsHelper(t.ExplicitMethod(i).Type(), imports)
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			m.importsHelper(t.EmbeddedType(i), imports)
		}
	}
}
