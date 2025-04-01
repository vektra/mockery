package template

import (
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
				return Data{Imports: []*Package{imprt}}
			},
			want: `x "xyz"`,
		},
		{
			name:       "syncPkgQualifier",
			inTemplate: "{{$.Imports.SyncPkgQualifier}}",
			dataInit: func() Data {
				return Data{Imports: []*Package{
					NewPackage(types.NewPackage("sync", "sync")),
					NewPackage(types.NewPackage("github.com/some/module", "module")),
				}}
			},
			want: "sync",
		},
		{
			name:       "syncPkgQualifier renamed",
			inTemplate: "{{$.Imports.SyncPkgQualifier}}",
			dataInit: func() Data {
				stdSync := NewPackage(types.NewPackage("sync", "sync"))
				stdSync.Alias = "stdSync"
				otherSyncPkg := NewPackage(types.NewPackage("github.com/someother/sync", "sync"))

				return Data{Imports: []*Package{stdSync, otherSyncPkg}}
			},
			want: "stdSync",
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
			name:       "MocksSomeMethod",
			inTemplate: "{{ .Mocks.MocksSomeMethod }}",
			dataInit: func() Data {
				// MethodData has to have at least 1 element to pass.
				return Data{Mocks: []MockData{{Methods: []MethodData{{}}}}}
			},
			want: "true",
		},
		{
			name:       "typeConstraint",
			inTemplate: "{{ (index .Mocks 0).TypeConstraintTest }}",
			dataInit: func() Data {
				return Data{Mocks: []MockData{{
					TypeParams: []TypeParamData{{
						ParamData: ParamData{
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
