package template

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/huandu/xstrings"
	"github.com/vektra/mockery/v3/pkg/registry"
)

// Template is the Moq template. It is capable of generating the Moq
// implementation for the given template.Data.
type Template struct {
	tmpl *template.Template
}

// New returns a new instance of Template.
func New(templateString string, name string) (Template, error) {
	mergedFuncMap := template.FuncMap{}
	for key, val := range StringManipulationFuncs {
		mergedFuncMap[key] = val
	}
	for key, val := range TemplateMockFuncs {
		mergedFuncMap[key] = val
	}

	tmpl, err := template.New(name).Funcs(mergedFuncMap).Parse(templateString)
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

var TemplateMockFuncs = template.FuncMap{
	"importStatement": func(imprt *registry.Package) string {
		if imprt.Alias == "" {
			return `"` + imprt.Path() + `"`
		}
		return imprt.Alias + ` "` + imprt.Path() + `"`
	},
	"syncPkgQualifier": func(imports []*registry.Package) string {
		for _, imprt := range imports {
			if imprt.Path() == "sync" {
				return imprt.Qualifier()
			}
		}

		return "sync"
	},
	"exported": exported,

	"mocksSomeMethod": func(mocks []MockData) bool {
		for _, m := range mocks {
			if len(m.Methods) > 0 {
				return true
			}
		}

		return false
	},
	"typeConstraintTest": func(m MockData) string {
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
}

var StringManipulationFuncs = template.FuncMap{
	// String inspection and manipulation. Note that the first argument is replaced
	// as the last argument in some functions in order to support chained
	// template pipelines.
	"contains":    func(substr string, s string) bool { return strings.Contains(s, substr) },
	"hasPrefix":   func(prefix string, s string) bool { return strings.HasPrefix(s, prefix) },
	"hasSuffix":   func(suffix string, s string) bool { return strings.HasSuffix(s, suffix) },
	"join":        func(sep string, elems []string) string { return strings.Join(elems, sep) },
	"replace":     func(old string, new string, n int, s string) string { return strings.Replace(s, old, new, n) },
	"replaceAll":  func(old string, new string, s string) string { return strings.ReplaceAll(s, old, new) },
	"split":       func(sep string, s string) []string { return strings.Split(s, sep) },
	"splitAfter":  func(sep string, s string) []string { return strings.SplitAfter(s, sep) },
	"splitAfterN": func(sep string, n int, s string) []string { return strings.SplitAfterN(s, sep, n) },
	"trim":        func(cutset string, s string) string { return strings.Trim(s, cutset) },
	"trimLeft":    func(cutset string, s string) string { return strings.TrimLeft(s, cutset) },
	"trimPrefix":  func(prefix string, s string) string { return strings.TrimPrefix(s, prefix) },
	"trimRight":   func(cutset string, s string) string { return strings.TrimRight(s, cutset) },
	"trimSpace":   strings.TrimSpace,
	"trimSuffix":  func(suffix string, s string) string { return strings.TrimSuffix(s, suffix) },
	"lower":       strings.ToLower,
	"upper":       strings.ToUpper,
	"camelcase":   xstrings.ToCamelCase,
	"snakecase":   xstrings.ToSnakeCase,
	"kebabcase":   xstrings.ToKebabCase,
	"firstLower":  xstrings.FirstRuneToLower,
	"firstUpper":  xstrings.FirstRuneToUpper,

	// Regular expression matching
	"matchString": regexp.MatchString,
	"quoteMeta":   regexp.QuoteMeta,

	// Filepath manipulation
	"base":  filepath.Base,
	"clean": filepath.Clean,
	"dir":   filepath.Dir,

	// Basic access to reading environment variables
	"expandEnv": os.ExpandEnv,
	"getenv":    os.Getenv,

	// Arithmetic
	"add": func(i1, i2 int) int { return i1 + i2 },
}
