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
