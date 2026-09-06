package internal

import (
	"bytes"
	"go/format"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/mockery/v3/template"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/tools/go/packages"
)

func TestFindPkgPath(t *testing.T) {
	pkgPath, err := findPkgPath("./fixtures")
	require.NoError(t, err)
	assert.NotEmpty(t, pkgPath)
}

func TestMatryerTemplateDataSchema(t *testing.T) {
	schema, err := gojsonschema.NewSchema(gojsonschema.NewStringLoader(templateMatryerJSONSchema))
	require.NoError(t, err)

	for _, tt := range []struct {
		name  string
		data  template.TemplateData
		valid bool
	}{
		{
			name: "embedded type",
			data: template.TemplateData{
				"struct-preamble": "gen.UnimplementedServer",
				"add-import":      []any{map[string]any{"name": "gen", "pkgPath": "example.com/gen"}},
			},
			valid: true,
		},
		{
			name:  "empty optional values",
			data:  template.TemplateData{"struct-preamble": "", "add-import": []any{}},
			valid: true,
		},
		{
			name: "preamble must be a string",
			data: template.TemplateData{"struct-preamble": []any{"gen.UnimplementedServer"}},
		},
		{
			name: "imports must be a list",
			data: template.TemplateData{"add-import": "example.com/gen"},
		},
		{
			name: "import requires a name",
			data: template.TemplateData{"add-import": []any{map[string]any{"pkgPath": "example.com/gen"}}},
		},
		{
			name: "import requires a path",
			data: template.TemplateData{"add-import": []any{map[string]any{"name": "gen"}}},
		},
		{
			name: "import rejects unknown fields",
			data: template.TemplateData{"add-import": []any{map[string]any{"name": "gen", "pkgPath": "example.com/gen", "alias": "other"}}},
		},
		{
			name: "import rejects empty path",
			data: template.TemplateData{"add-import": []any{map[string]any{"name": "gen", "pkgPath": ""}}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.data.VerifyJSONSchema(t.Context(), schema)
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, template.ErrTemplateDataSchemaValidation)
			}
		})
	}
}

func TestMatryerAdditionalImports(t *testing.T) {
	for _, tt := range []struct {
		name         string
		firstAlias   string
		secondAlias  string
		secondPath   string
		methodImport bool
		wantError    string
	}{
		{name: "shared import", firstAlias: "gen", secondAlias: "gen", secondPath: "example.com/generated"},
		{name: "distinct imports", firstAlias: "gen", secondAlias: "other", secondPath: "example.com/other"},
		{name: "alias collision across interfaces", firstAlias: "gen", secondAlias: "gen", secondPath: "example.com/other", wantError: "name already used by import"},
		{name: "different aliases for the same path", firstAlias: "gen", secondAlias: "other", secondPath: "example.com/generated", wantError: "already imported as"},
		{name: "alias collision with a method import", firstAlias: "context", secondAlias: "gen", secondPath: "example.com/generated", methodImport: true, wantError: `name already used by import "context"`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, err := template.NewRegistry(&packages.Package{Name: "source", PkgPath: "example.com/source"}, "example.com/mocks", false)
			require.NoError(t, err)
			if tt.methodImport {
				r.AddImport("context", "context")
			}
			data := template.Data{
				PkgName:      "mocks",
				Registry:     r,
				TemplateData: template.TemplateData{"skip-ensure": true},
				Interfaces: template.Interfaces{
					{Name: "First", StructName: "MockFirst", TemplateData: template.TemplateData{
						"skip-ensure":     true,
						"struct-preamble": tt.firstAlias + ".FirstBase",
						"add-import":      []any{map[string]any{"name": tt.firstAlias, "pkgPath": "example.com/generated"}},
					}},
					{Name: "Second", StructName: "MockSecond", TemplateData: template.TemplateData{
						"skip-ensure":     true,
						"struct-preamble": tt.secondAlias + ".SecondBase",
						"add-import":      []any{map[string]any{"name": tt.secondAlias, "pkgPath": tt.secondPath}},
					}},
				},
			}
			tmpl, err := template.New(templateMatryer, "matryer")
			require.NoError(t, err)
			var out bytes.Buffer
			err = tmpl.Execute(&out, data)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			_, err = format.Source(out.Bytes())
			require.NoError(t, err)
			require.Contains(t, out.String(), tt.firstAlias+".FirstBase")
			require.Contains(t, out.String(), tt.secondAlias+".SecondBase")
			if tt.secondPath == "example.com/generated" {
				require.Len(t, r.Imports(), 1)
			} else {
				require.Len(t, r.Imports(), 2)
			}
		})
	}
}
