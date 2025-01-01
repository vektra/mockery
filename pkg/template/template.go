package template

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	_ "embed"

	"github.com/huandu/xstrings"
	"github.com/vektra/mockery/v3/pkg/registry"
	"github.com/vektra/mockery/v3/pkg/stackerr"
)

// Template is the Moq template. It is capable of generating the Moq
// implementation for the given template.Data.
type Template struct {
	tmpl *template.Template
}

var (
	//go:embed moq.templ
	templateMoq string
	//go:embed mockery.templ
	templateMockery string
)

var styleTemplates = map[string]string{
	"moq":     templateMoq,
	"mockery": templateMockery,
}

// New returns a new instance of Template.
func New(style string) (Template, error) {
	templateString, styleExists := styleTemplates[style]
	if !styleExists {
		return Template{}, stackerr.NewStackErrf(nil, "style %s does not exist", style)
	}

	tmpl, err := template.New(style).Funcs(templateFuncs).Parse(templateString)
	if err != nil {
		return Template{}, err
	}

	return Template{tmpl: tmpl}, nil
}

// Execute generates and writes the Moq implementation for the given
// data.
func (t Template) Execute(w io.Writer, data Data) error {
	return t.tmpl.Execute(w, data)
}

// This list comes from the golint codebase. Golint will complain about any of
// these being mixed-case, like "Id" instead of "ID".
var golintInitialisms = []string{
	"ACL", "API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID", "IP", "JSON", "LHS",
	"QPS", "RAM", "RHS", "RPC", "SLA", "SMTP", "SQL", "SSH", "TCP", "TLS", "TTL", "UDP", "UI", "UID", "UUID", "URI",
	"URL", "UTF8", "VM", "XML", "XMPP", "XSRF", "XSS",
}

func exported(s string) string {
	if s == "" {
		return ""
	}
	for _, initialism := range golintInitialisms {
		if strings.ToUpper(s) == initialism {
			return initialism
		}
	}
	return strings.ToUpper(s[0:1]) + s[1:]
}

var templateFuncs = template.FuncMap{
	"ImportStatement": func(imprt *registry.Package) string {
		if imprt.Alias == "" {
			return `"` + imprt.Path() + `"`
		}
		return imprt.Alias + ` "` + imprt.Path() + `"`
	},
	"SyncPkgQualifier": func(imports []*registry.Package) string {
		for _, imprt := range imports {
			if imprt.Path() == "sync" {
				return imprt.Qualifier()
			}
		}

		return "sync"
	},
	"Exported": exported,

	"MocksSomeMethod": func(mocks []MockData) bool {
		for _, m := range mocks {
			if len(m.Methods) > 0 {
				return true
			}
		}

		return false
	},
	"TypeConstraintTest": func(m MockData) string {
		if len(m.TypeParams) == 0 {
			return ""
		}
		s := "["
		for idx, param := range m.TypeParams {
			if idx != 0 {
				s += ", "
			}
			s += exported(param.Name())
			s += " "
			s += param.TypeString()
		}
		s += "]"
		return s
	},
	// String inspection and manipulation. Note that the first argument is replaced
	// as the last argument in some functions in order to support chained
	// template pipelines.
	"Contains":    func(substr string, s string) bool { return strings.Contains(s, substr) },
	"HasPrefix":   func(prefix string, s string) bool { return strings.HasPrefix(s, prefix) },
	"HasSuffix":   func(suffix string, s string) bool { return strings.HasSuffix(s, suffix) },
	"Join":        func(sep string, elems []string) string { return strings.Join(elems, sep) },
	"Replace":     func(old string, new string, n int, s string) string { return strings.Replace(s, old, new, n) },
	"ReplaceAll":  func(old string, new string, s string) string { return strings.ReplaceAll(s, old, new) },
	"Split":       func(sep string, s string) []string { return strings.Split(s, sep) },
	"SplitAfter":  func(sep string, s string) []string { return strings.SplitAfter(s, sep) },
	"SplitAfterN": func(sep string, n int, s string) []string { return strings.SplitAfterN(s, sep, n) },
	"Trim":        func(cutset string, s string) string { return strings.Trim(s, cutset) },
	"TrimLeft":    func(cutset string, s string) string { return strings.TrimLeft(s, cutset) },
	"TrimPrefix":  func(prefix string, s string) string { return strings.TrimPrefix(s, prefix) },
	"TrimRight":   func(cutset string, s string) string { return strings.TrimRight(s, cutset) },
	"TrimSpace":   strings.TrimSpace,
	"TrimSuffix":  func(suffix string, s string) string { return strings.TrimSuffix(s, suffix) },
	"Lower":       strings.ToLower,
	"Upper":       strings.ToUpper,
	"Camelcase":   xstrings.ToCamelCase,
	"Snakecase":   xstrings.ToSnakeCase,
	"Kebabcase":   xstrings.ToKebabCase,
	"FirstLower":  xstrings.FirstRuneToLower,
	"FirstUpper":  xstrings.FirstRuneToUpper,

	// Regular expression matching
	"MatchString": regexp.MatchString,
	"QuoteMeta":   regexp.QuoteMeta,

	// Filepath manipulation
	"Base":  filepath.Base,
	"Clean": filepath.Clean,
	"Dir":   filepath.Dir,

	// Basic access to reading environment variables
	"ExpandEnv": os.ExpandEnv,
	"Getenv":    os.Getenv,

	// Arithmetic
	"Add": func(i1, i2 int) int { return i1 + i2 },
}
