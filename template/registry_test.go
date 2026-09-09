package template

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

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
