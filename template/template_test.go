package template

import (
	"context"
	"go/types"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateMockFuncs(t *testing.T) {
	tests := []struct {
		name       string
		inTemplate string
		dataInit   func() Data
		want       string
	}{
		{
			name:       "importStatement",
			inTemplate: "{{- range .Imports}}{{ .ImportStatement }}{{- end}}",
			dataInit: func() Data {
				imprt := NewPackage(types.NewPackage("xyz", "xyz"))
				imprt.Alias = "x"
				registry, err := NewRegistry(nil, "", false)
				require.NoError(t, err)
				registry.addImport(context.Background(), imprt.pkg)

				return Data{Registry: registry}
			},
			want: `"xyz"`,
		},
		{
			name:       "PkgQualifier",
			inTemplate: `{{$.Imports.PkgQualifier "sync"}}`,
			dataInit: func() Data {
				registry, err := NewRegistry(nil, "", false)
				require.NoError(t, err)
				registry.addImport(context.Background(), NewPackage(types.NewPackage("sync", "sync")).pkg)
				registry.addImport(context.Background(), NewPackage(types.NewPackage("github.com/some/module", "module")).pkg)

				return Data{Registry: registry}
			},
			want: "sync",
		},
		{
			name:       "PkgQualifier conflicting pkg names",
			inTemplate: `{{$.Imports.PkgQualifier "github.com/someother/sync"}}`,
			dataInit: func() Data {
				registry, err := NewRegistry(nil, "", false)
				require.NoError(t, err)
				registry.AddImport("sync", "sync")
				registry.AddImport("sync", "github.com/someother/sync")

				return Data{Registry: registry}
			},
			want: "sync0",
		},
		{
			name:       "exported empty",
			inTemplate: "{{exported .TemplateData.var}}",
			dataInit:   func() Data { return Data{TemplateData: map[string]any{"var": ""}} },
			want:       "",
		},
		{
			name:       "exported var",
			inTemplate: "{{exported .TemplateData.var}}",
			dataInit:   func() Data { return Data{TemplateData: map[string]any{"var": "someVar"}} },
			want:       "SomeVar",
		},
		{
			name:       "exported acronym",
			inTemplate: "{{exported .TemplateData.var}}",
			dataInit:   func() Data { return Data{TemplateData: map[string]any{"var": "sql"}} },
			want:       "SQL",
		},
		{
			name:       "ImplementsSomeMethod",
			inTemplate: "{{ .Interfaces.ImplementsSomeMethod }}",
			dataInit: func() Data {
				// MethodData has to have at least 1 element to pass.
				return Data{Interfaces: []Interface{{Methods: []Method{{}}}}}
			},
			want: "true",
		},
		{
			name:       "typeConstraint",
			inTemplate: "{{ (index .Interfaces 0).TypeConstraintTest }}",
			dataInit: func() Data {
				return Data{Interfaces: []Interface{{
					TypeParams: []TypeParam{{
						Param: Param{
							Var: &Var{Name: "t", typ: &types.Slice{}},
						},
					}},
				}}}
			},
			want: "[T []<nil>]",
		},
		{
			name:       "readFile",
			inTemplate: "{{readFile .TemplateData.f}}",
			dataInit: func() Data {
				f, err := os.CreateTemp(".", "readFileTest")
				require.NoError(t, err)

				t.Cleanup(func() {
					os.Remove(f.Name())
				})

				_, err = f.WriteString("content")
				require.NoError(t, err)

				return Data{TemplateData: map[string]any{"f": f.Name()}}
			},
			want: "content",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tt, err := New(tc.inTemplate, tc.name)
			require.NoError(t, err)

			var sb strings.Builder
			err = tt.Execute(&sb, tc.dataInit())
			require.NoError(t, err)
			assert.Equal(t, tc.want, sb.String())
		})
	}
}
