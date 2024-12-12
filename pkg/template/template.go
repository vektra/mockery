package template

import (
	"io"
	"strings"
	"text/template"

	_ "embed"

	"github.com/vektra/mockery/v2/pkg/registry"
	"github.com/vektra/mockery/v2/pkg/stackerr"
)

// Template is the Moq template. It is capable of generating the Moq
// implementation for the given template.Data.
type Template struct {
	tmpl *template.Template
}

//go:embed moq.templ
var templateMoq string

var styleTemplates = map[string]string{
	"moq": templateMoq,
}

// New returns a new instance of Template.
func New(style string) (Template, error) {
	templateString, styleExists := styleTemplates[style]
	if !styleExists {
		return Template{}, stackerr.NewStackErrf(nil, "style %s does not exist", style)
	}

	tmpl, err := template.New("moq").Funcs(templateFuncs).Parse(templateString)
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
	"TypeInstantiation": func(m MockData) string {
		if len(m.TypeParams) == 0 {
			return ""
		}
		s := "["
		for idx, param := range m.TypeParams {
			if idx != 0 {
				s += ", "
			}
			s += exported(param.Name())
		}
		s += "]"
		return s
	},
}
