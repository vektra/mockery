package template

import (
	"go/types"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateStringFuncs(t *testing.T) {
	// For env tests
	os.Setenv("MOCKERY_TEST_ENV", "TEST")

	tests := []struct {
		name       string
		inTemplate string
		inData     map[string]any
		want       string
	}{
		{
			name:       "contains",
			inTemplate: "{{contains .TemplateData.sub .TemplateData.str}}",
			inData:     map[string]any{"str": "golang", "sub": "go"},
			want:       "true",
		},
		{
			name:       "hasPrefix",
			inTemplate: "{{hasPrefix .TemplateData.pre .TemplateData.str}}",
			inData:     map[string]any{"str": "golang", "pre": "go"},
			want:       "true",
		},
		{
			name:       "hasSuffix",
			inTemplate: "{{hasSuffix .TemplateData.suf .TemplateData.str}}",
			inData:     map[string]any{"str": "golang", "suf": "lang"},
			want:       "true",
		},
		{
			name:       "join",
			inTemplate: "{{join .TemplateData.sep .TemplateData.elems}}",
			inData:     map[string]any{"elems": []string{"1", "2", "3"}, "sep": ","},
			want:       "1,2,3",
		},
		{
			name:       "replace",
			inTemplate: "{{replace .TemplateData.old .TemplateData.new .TemplateData.n .TemplateData.s}}",
			inData:     map[string]any{"old": "old", "new": "new", "n": 2, "s": "oldoldold"},
			want:       "newnewold",
		},
		{
			name:       "replaceAll",
			inTemplate: "{{replaceAll .TemplateData.old .TemplateData.new .TemplateData.s}}",
			inData:     map[string]any{"old": "old", "new": "new", "s": "oldoldold"},
			want:       "newnewnew",
		},

		// String splitting
		{
			name:       "split",
			inTemplate: "{{split .TemplateData.sep .TemplateData.s}}",
			inData:     map[string]any{"s": "a,b,c", "sep": ","},
			want:       "[a b c]",
		},
		{
			name:       "splitAfter",
			inTemplate: "{{splitAfter .TemplateData.sep .TemplateData.s}}",
			inData:     map[string]any{"s": "a,b,c", "sep": ","},
			want:       "[a, b, c]",
		},
		{
			name:       "splitAfterN",
			inTemplate: "{{splitAfterN .TemplateData.sep .TemplateData.n .TemplateData.s}}",
			inData:     map[string]any{"s": "a,b,c,d", "sep": ",", "n": 2},
			want:       "[a, b,c,d]",
		},

		// Trimming functions
		{
			name:       "trim",
			inTemplate: "{{trim .TemplateData.cutset .TemplateData.s}}",
			inData:     map[string]any{"s": "---hello---", "cutset": "-"},
			want:       "hello",
		},
		{
			name:       "trimLeft",
			inTemplate: "{{trimLeft .TemplateData.cutset .TemplateData.s}}",
			inData:     map[string]any{"s": "---hello---", "cutset": "-"},
			want:       "hello---",
		},
		{
			name:       "trimRight",
			inTemplate: "{{trimRight .TemplateData.cutset .TemplateData.s}}",
			inData:     map[string]any{"s": "---hello---", "cutset": "-"},
			want:       "---hello",
		},
		{
			name:       "trimPrefix",
			inTemplate: "{{trimPrefix .TemplateData.prefix .TemplateData.s}}",
			inData:     map[string]any{"s": "prefix_text", "prefix": "prefix_"},
			want:       "text",
		},
		{
			name:       "trimSuffix",
			inTemplate: "{{trimSuffix .TemplateData.suffix .TemplateData.s}}",
			inData:     map[string]any{"s": "text_suffix", "suffix": "_suffix"},
			want:       "text",
		},
		{
			name:       "trimSpace",
			inTemplate: "{{trimSpace .TemplateData.s}}",
			inData:     map[string]any{"s": "   hello world   "},
			want:       "hello world",
		},

		// Casing functions
		{
			name:       "lower",
			inTemplate: "{{lower .TemplateData.s}}",
			inData:     map[string]any{"s": "GoLang"},
			want:       "golang",
		},
		{
			name:       "upper",
			inTemplate: "{{upper .TemplateData.s}}",
			inData:     map[string]any{"s": "golang"},
			want:       "GOLANG",
		},
		{
			name:       "camelcase",
			inTemplate: "{{camelcase .TemplateData.s}}",
			inData:     map[string]any{"s": "hello_world"},
			want:       "helloWorld",
		},
		{
			name:       "snakecase",
			inTemplate: "{{snakecase .TemplateData.s}}",
			inData:     map[string]any{"s": "HelloWorld"},
			want:       "hello_world",
		},
		{
			name:       "kebabcase",
			inTemplate: "{{kebabcase .TemplateData.s}}",
			inData:     map[string]any{"s": "HelloWorld"},
			want:       "hello-world",
		},
		{
			name:       "firstLower",
			inTemplate: "{{firstLower .TemplateData.s}}",
			inData:     map[string]any{"s": "GoLang"},
			want:       "goLang",
		},
		{
			name:       "firstUpper",
			inTemplate: "{{firstUpper .TemplateData.s}}",
			inData:     map[string]any{"s": "golang"},
			want:       "Golang",
		},

		// Regex functions
		{
			name:       "matchString",
			inTemplate: "{{matchString .TemplateData.pattern .TemplateData.s}}",
			inData:     map[string]any{"pattern": "go.*", "s": "golang"},
			want:       "true",
		},
		{
			name:       "quoteMeta",
			inTemplate: "{{quoteMeta .TemplateData.s}}",
			inData:     map[string]any{"s": "1+1=2"},
			want:       `1\+1=2`,
		},

		// Filepath manipulation
		{
			name:       "base",
			inTemplate: "{{base .TemplateData.s}}",
			inData:     map[string]any{"s": "/home/user/file.txt"},
			want:       "file.txt",
		},
		{
			name:       "clean",
			inTemplate: "{{clean .TemplateData.s}}",
			inData:     map[string]any{"s": "/home/user/../file.txt"},
			want:       "/home/file.txt",
		},
		{
			name:       "dir",
			inTemplate: "{{dir .TemplateData.s}}",
			inData:     map[string]any{"s": "/home/user/file.txt"},
			want:       "/home/user",
		},

		// Environment variables
		{
			name:       "getenv",
			inTemplate: "{{getenv .TemplateData.s}}",
			inData:     map[string]any{"s": "MOCKERY_TEST_ENV"},
			want:       "TEST",
		},
		{
			name:       "expandEnv",
			inTemplate: "{{expandEnv .TemplateData.s}}",
			inData:     map[string]any{"s": "${MOCKERY_TEST_ENV}"},
			want:       "TEST",
		},

		// Arithmetic
		{
			name:       "add",
			inTemplate: "{{add .TemplateData.i1 .TemplateData.i2}}",
			inData:     map[string]any{"i1": 5, "i2": 10},
			want:       "15",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tt, err := New(tc.inTemplate, tc.name)
			require.NoError(t, err)

			var sb strings.Builder
			err = tt.Execute(&sb, Data{TemplateData: tc.inData})
			require.NoError(t, err)
			assert.Equal(t, tc.want, sb.String())
		})
	}
}

func TestTemplateMockFuncs(t *testing.T) {
	tests := []struct {
		name       string
		inTemplate string
		dataInit   func() Data
		want       string
	}{
		{
			name:       "importStatement",
			inTemplate: "{{- range .Imports}}{{. | importStatement}}{{- end}}",
			dataInit: func() Data {
				imprt := NewPackage(types.NewPackage("xyz", "xyz"))
				imprt.Alias = "x"
				return Data{Imports: []*Package{imprt}}
			},
			want: `x "xyz"`,
		},
		{
			name:       "syncPkgQualifier",
			inTemplate: "{{$.Imports | syncPkgQualifier}}",
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
			inTemplate: "{{$.Imports | syncPkgQualifier}}",
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
			name:       "mocksSomeMethod",
			inTemplate: "{{mocksSomeMethod .Mocks}}",
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
