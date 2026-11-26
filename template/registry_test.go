package template

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

// A generic alias of a generic interface declares its own type parameters, which the
// mock must carry (`type L[T any] = impl.L[T]`).
func TestRegistryLookupInterfaceGenericAlias(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "lib.go", `package lib
type Limiter[T any] interface{ Permit(item T) error }
type Alias[T any] = Limiter[T]
type Plain interface{ M() }
`, 0)
	require.NoError(t, err)
	pkg, err := (&types.Config{}).Check("example.com/lib", fset, []*ast.File{f}, nil)
	require.NoError(t, err)
	registry, err := NewRegistry(&packages.Package{Name: pkg.Name(), PkgPath: pkg.Path(), Types: pkg}, "example.com/output", false)
	require.NoError(t, err)

	for name, want := range map[string]int{"Limiter": 1, "Alias": 1, "Plain": 0} {
		iface, tparams, err := registry.LookupInterface(name)
		require.NoError(t, err, name)
		require.NotNil(t, iface, name)
		require.Equal(t, want, tparams.Len(), name)
	}
}

func TestRegistryAddImportPackageScopeConflicts(t *testing.T) {
	for _, tt := range []struct {
		name          string
		inPackage     bool
		reservedNames []string
		wantAlias     string
		wantQualifier string
	}{
		{
			name:          "in-package declaration conflict",
			inPackage:     true,
			reservedNames: []string{"dep"},
			wantAlias:     "dep0",
			wantQualifier: "dep0",
		},
		{
			name:          "out-of-package ignores source scope",
			inPackage:     false,
			reservedNames: []string{"dep"},
			wantQualifier: "dep",
		},
		{
			name:          "in-package conflict advances past numbered declaration",
			inPackage:     true,
			reservedNames: []string{"dep", "dep0"},
			wantAlias:     "dep1",
			wantQualifier: "dep1",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srcTypes := types.NewPackage("example.com/consumer", "dep")
			for _, name := range tt.reservedNames {
				require.Nil(t, srcTypes.Scope().Insert(types.NewTypeName(
					token.NoPos,
					srcTypes,
					name,
					types.NewStruct(nil, nil),
				)))
			}
			srcPkg := &packages.Package{
				Name:    srcTypes.Name(),
				PkgPath: srcTypes.Path(),
				Types:   srcTypes,
			}
			registry, err := NewRegistry(srcPkg, "example.com/output", tt.inPackage)
			require.NoError(t, err)

			imprt := registry.AddImport("dep", "example.com/lib/dep")
			require.Equal(t, tt.wantAlias, imprt.Alias)
			require.Equal(t, tt.wantQualifier, imprt.Qualifier())
		})
	}
}

// goimports rewrites an import to add an explicit alias whenever the declared
// package name differs from the name it infers from the import path. The
// generated file must already contain that alias, otherwise running goimports
// over a freshly generated mock produces a diff.
func TestRegistryAddImportEmitsGoimportsAlias(t *testing.T) {
	for _, tt := range []struct {
		name      string
		pkgName   string
		pkgPath   string
		wantAlias string
		wantStmt  string
	}{
		{
			name:      "package name differs from directory",
			pkgName:   "bar",
			pkgPath:   "mockery/foo",
			wantAlias: "bar",
			wantStmt:  `bar "mockery/foo"`,
		},
		{
			name:      "major version suffix is not part of the name",
			pkgName:   "semver",
			pkgPath:   "github.com/Masterminds/semver/v3",
			wantAlias: "",
			wantStmt:  `"github.com/Masterminds/semver/v3"`,
		},
		{
			name:      "non-identifier suffix is trimmed",
			pkgName:   "yaml",
			pkgPath:   "gopkg.in/yaml.v3",
			wantAlias: "",
			wantStmt:  `"gopkg.in/yaml.v3"`,
		},
		{
			name:      "go- prefix is trimmed",
			pkgName:   "cmd",
			pkgPath:   "example.com/go-cmd",
			wantAlias: "",
			wantStmt:  `"example.com/go-cmd"`,
		},
		{
			name:      "matching names need no alias",
			pkgName:   "dep",
			pkgPath:   "example.com/lib/dep",
			wantAlias: "",
			wantStmt:  `"example.com/lib/dep"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewRegistry(nil, "example.com/output", false)
			require.NoError(t, err)

			imprt := r.AddImport(tt.pkgName, tt.pkgPath)
			require.Equal(t, tt.wantAlias, imprt.Alias)
			require.Equal(t, tt.wantStmt, imprt.ImportStatement())
		})
	}
}

func TestImportPathAssumedName(t *testing.T) {
	for _, tt := range []struct {
		path string
		want string
	}{
		{path: "mockery/foo", want: "foo"},
		{path: "github.com/Masterminds/semver/v3", want: "semver"},
		{path: "github.com/foo/bar/v12", want: "bar"},
		{path: "gopkg.in/yaml.v3", want: "yaml"},
		{path: "example.com/go-cmd", want: "cmd"},
		{path: "example.com/lib/dep", want: "dep"},
		{path: "stdlib", want: "stdlib"},
	} {
		require.Equal(t, tt.want, importPathAssumedName(tt.path), tt.path)
	}
}

func TestRegistryAddImportWithAlias(t *testing.T) {
	for _, tt := range []struct {
		name      string
		alias     string
		path      string
		setup     func(*Registry)
		wantError string
	}{
		{name: "explicit alias", alias: "service", path: "example.com/generated"},
		{name: "empty alias", path: "example.com/generated", wantError: "invalid import name"},
		{name: "keyword", alias: "type", path: "example.com/generated", wantError: "invalid import name"},
		{name: "blank import", alias: "_", path: "example.com/generated", wantError: "invalid import name"},
		{name: "dot import", alias: ".", path: "example.com/generated", wantError: "invalid import name"},
		{name: "empty path", alias: "service", wantError: "non-empty package path"},
		{name: "self import", alias: "source", path: "example.com/source", wantError: "cannot import source package"},
		{name: "package declaration conflict", alias: "Taken", path: "example.com/generated", wantError: "name already declared"},
		{
			name:  "existing method import",
			alias: "context",
			path:  "context",
			setup: func(r *Registry) { r.AddImport("context", "context") },
		},
		{
			name:      "conflict with method import",
			alias:     "context",
			path:      "example.com/generated",
			setup:     func(r *Registry) { r.AddImport("context", "context") },
			wantError: `name already used by import "context"`,
		},
		{
			name:      "same path with a different name",
			alias:     "ctx",
			path:      "context",
			setup:     func(r *Registry) { r.AddImport("context", "context") },
			wantError: `already imported as "context"`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srcTypes := types.NewPackage("example.com/source", "source")
			srcTypes.Scope().Insert(types.NewTypeName(token.NoPos, srcTypes, "Taken", types.NewStruct(nil, nil)))
			r, err := NewRegistry(&packages.Package{PkgPath: srcTypes.Path(), Name: srcTypes.Name(), Types: srcTypes}, srcTypes.Path(), true)
			require.NoError(t, err)
			if tt.setup != nil {
				tt.setup(r)
			}
			before := r.Imports()
			imprt, err := r.AddImportWithAlias(tt.alias, tt.path)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				require.Nil(t, imprt)
				require.Equal(t, before, r.Imports())
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.alias, imprt.Qualifier())
			require.Equal(t, tt.path, imprt.Path())
			if tt.setup == nil {
				require.Equal(t, `service "example.com/generated"`, imprt.ImportStatement())
			}
			again, err := r.AddImportWithAlias(tt.alias, tt.path)
			require.NoError(t, err)
			require.Same(t, imprt, again)
			require.Same(t, imprt, r.AddImport("generated", tt.path))
			require.Len(t, r.Imports(), 1)
		})
	}
}
