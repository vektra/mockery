// Package template provides data and functionality for rendering templates using mockery.
package template

import (
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/vektra/mockery/v3/shared"
)

// Template is the Moq template. It is capable of generating the Moq
// implementation for the given template.Data.
type Template struct {
	tmpl *template.Template
}

// New returns a new instance of Template.
func New(templateString string, name string) (Template, error) {
	mergedFuncMap := template.FuncMap{}
	for key, val := range shared.StringManipulationFuncs {
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
	"importStatement": func(imprt *Package) string {
		if imprt.Alias == "" {
			return `"` + imprt.Path() + `"`
		}
		return imprt.Alias + ` "` + imprt.Path() + `"`
	},
	"syncPkgQualifier": func(imports []*Package) string {
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
	"readFile": func(path string) string {
		if path == "" {
			return ""
		}
		fileBytes, err := os.ReadFile(path)
		if err != nil {
			panic(err.Error())
		}
		return string(fileBytes)
	},
}
