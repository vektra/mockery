package template_funcs

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Data struct {
	TemplateData map[string]any
}

func TestTemplateStringFuncs(t *testing.T) {
	// For env tests
	os.Setenv("MOCKERY_TEST_ENV", "TEST")

	tests := []struct {
		name      string
		template  string
		data      map[string]any
		want      string
		wantRegex string
	}{
		{
			name:     "contains",
			template: "{{contains .TemplateData.sub .TemplateData.str}}",
			data:     map[string]any{"str": "golang", "sub": "go"},
			want:     "true",
		},
		{
			name:     "hasPrefix",
			template: "{{hasPrefix .TemplateData.pre .TemplateData.str}}",
			data:     map[string]any{"str": "golang", "pre": "go"},
			want:     "true",
		},
		{
			name:     "hasSuffix",
			template: "{{hasSuffix .TemplateData.suf .TemplateData.str}}",
			data:     map[string]any{"str": "golang", "suf": "lang"},
			want:     "true",
		},
		{
			name:     "join",
			template: "{{join .TemplateData.sep .TemplateData.elems}}",
			data:     map[string]any{"elems": []string{"1", "2", "3"}, "sep": ","},
			want:     "1,2,3",
		},
		{
			name:     "replace",
			template: "{{replace .TemplateData.old .TemplateData.new .TemplateData.n .TemplateData.s}}",
			data:     map[string]any{"old": "old", "new": "new", "n": 2, "s": "oldoldold"},
			want:     "newnewold",
		},
		{
			name:     "replaceAll",
			template: "{{replaceAll .TemplateData.old .TemplateData.new .TemplateData.s}}",
			data:     map[string]any{"old": "old", "new": "new", "s": "oldoldold"},
			want:     "newnewnew",
		},

		// String splitting
		{
			name:     "split",
			template: "{{split .TemplateData.sep .TemplateData.s}}",
			data:     map[string]any{"s": "a,b,c", "sep": ","},
			want:     "[a b c]",
		},
		{
			name:     "splitAfter",
			template: "{{splitAfter .TemplateData.sep .TemplateData.s}}",
			data:     map[string]any{"s": "a,b,c", "sep": ","},
			want:     "[a, b, c]",
		},
		{
			name:     "splitAfterN",
			template: "{{splitAfterN .TemplateData.sep .TemplateData.n .TemplateData.s}}",
			data:     map[string]any{"s": "a,b,c,d", "sep": ",", "n": 2},
			want:     "[a, b,c,d]",
		},

		// Trimming functions
		{
			name:     "trim",
			template: "{{trim .TemplateData.cutset .TemplateData.s}}",
			data:     map[string]any{"s": "---hello---", "cutset": "-"},
			want:     "hello",
		},
		{
			name:     "trimLeft",
			template: "{{trimLeft .TemplateData.cutset .TemplateData.s}}",
			data:     map[string]any{"s": "---hello---", "cutset": "-"},
			want:     "hello---",
		},
		{
			name:     "trimRight",
			template: "{{trimRight .TemplateData.cutset .TemplateData.s}}",
			data:     map[string]any{"s": "---hello---", "cutset": "-"},
			want:     "---hello",
		},
		{
			name:     "trimPrefix",
			template: "{{trimPrefix .TemplateData.prefix .TemplateData.s}}",
			data:     map[string]any{"s": "prefix_text", "prefix": "prefix_"},
			want:     "text",
		},
		{
			name:     "trimSuffix",
			template: "{{trimSuffix .TemplateData.suffix .TemplateData.s}}",
			data:     map[string]any{"s": "text_suffix", "suffix": "_suffix"},
			want:     "text",
		},
		{
			name:     "trimSpace",
			template: "{{trimSpace .TemplateData.s}}",
			data:     map[string]any{"s": "   hello world   "},
			want:     "hello world",
		},

		// Casing functions
		{
			name:     "lower",
			template: "{{lower .TemplateData.s}}",
			data:     map[string]any{"s": "GoLang"},
			want:     "golang",
		},
		{
			name:     "upper",
			template: "{{upper .TemplateData.s}}",
			data:     map[string]any{"s": "golang"},
			want:     "GOLANG",
		},
		{
			name:     "camelcase",
			template: "{{camelcase .TemplateData.s}}",
			data:     map[string]any{"s": "hello_world"},
			want:     "helloWorld",
		},
		{
			name:     "snakecase",
			template: "{{snakecase .TemplateData.s}}",
			data:     map[string]any{"s": "HelloWorld"},
			want:     "hello_world",
		},
		{
			name:     "kebabcase",
			template: "{{kebabcase .TemplateData.s}}",
			data:     map[string]any{"s": "HelloWorld"},
			want:     "hello-world",
		},
		{
			name:     "firstLower",
			template: "{{firstLower .TemplateData.s}}",
			data:     map[string]any{"s": "GoLang"},
			want:     "goLang",
		},
		{
			name:     "firstUpper",
			template: "{{firstUpper .TemplateData.s}}",
			data:     map[string]any{"s": "golang"},
			want:     "Golang",
		},

		// Regex functions
		{
			name:     "matchString",
			template: "{{matchString .TemplateData.pattern .TemplateData.s}}",
			data:     map[string]any{"pattern": "go.*", "s": "golang"},
			want:     "true",
		},
		{
			name:     "quoteMeta",
			template: "{{quoteMeta .TemplateData.s}}",
			data:     map[string]any{"s": "1+1=2"},
			want:     `1\+1=2`,
		},

		// Filepath manipulation
		{
			name:     "base",
			template: "{{base .TemplateData.s}}",
			data:     map[string]any{"s": "/home/user/file.txt"},
			want:     "file.txt",
		},
		{
			name:     "clean",
			template: "{{clean .TemplateData.s}}",
			data:     map[string]any{"s": "/home/user/../file.txt"},
			want:     "/home/file.txt",
		},
		{
			name:     "dir",
			template: "{{dir .TemplateData.s}}",
			data:     map[string]any{"s": "/home/user/file.txt"},
			want:     "/home/user",
		},

		// Environment variables
		{
			name:     "getenv",
			template: "{{getenv .TemplateData.s}}",
			data:     map[string]any{"s": "MOCKERY_TEST_ENV"},
			want:     "TEST",
		},
		{
			name:     "expandEnv",
			template: "{{expandEnv .TemplateData.s}}",
			data:     map[string]any{"s": "${MOCKERY_TEST_ENV}"},
			want:     "TEST",
		},

		// Arithmetic
		{
			name:     "add",
			template: "{{add .TemplateData.i1 .TemplateData.i2}}",
			data:     map[string]any{"i1": 5, "i2": 10},
			want:     "15",
		},
		{
			name:     "decr",
			template: "{{decr 15}}",
			want:     "14",
		},
		{
			name:     "div",
			template: "{{div 28 7}}",
			want:     "4",
		},
		{
			name:     "incr",
			template: "{{incr 1}}",
			want:     "2",
		},
		{
			name:     "min",
			template: "{{min 2 4 6}}",
			want:     "2",
		},
		{
			name:     "mod",
			template: "{{mod 5 2}}",
			want:     "1",
		},
		{
			name:     "mul",
			template: "{{mul 5 2}}",
			want:     "10",
		},
		{
			name:     "sub",
			template: "{{sub 5 2}}",
			want:     "3",
		},
		{
			name:     "ceil",
			template: "{{ceil 1.71}}",
			want:     "2",
		},
		{
			name:     "floor",
			template: "{{floor 1.71}}",
			want:     "1",
		},
		{
			name:     "round 1.6",
			template: "{{round 1.6}}",
			want:     "2",
		},
		{
			name:     "round 1.4",
			template: "{{round 1.4}}",
			want:     "1",
		},
		{
			name:      "randInt",
			template:  "{{randInt}}",
			wantRegex: "%d",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmpl, err := template.New(tc.name).Funcs(FuncMap).Parse(tc.template)
			require.NoError(t, err)
			var sb strings.Builder
			err = tmpl.Execute(&sb, Data{TemplateData: tc.data})
			require.NoError(t, err)

			if tc.wantRegex == "" {
				assert.Equal(t, tc.want, sb.String())
			} else {
				re, err := regexp.Compile(tc.wantRegex)
				require.NoError(t, err)
				re.Match([]byte(sb.String()))
			}
		})
	}
}
