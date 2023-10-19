package template

import (
	"io"
	"text/template"
)

// Template is the Moq template. It is capable of generating the Moq
// implementation for the given template.Data.
type Template struct {
	tmpl *template.Template
}

//go:embed moq.templ
var moqTemplate string

var styleMap = map[string]string{
	"moq": moqTemplate,
}

// New returns a new instance of Template.
func New(style string) (Template, error) {
	tmpl, err := template.New("moq").Funcs(templateFuncs).Parse(styleMap[style])
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

var templateFuncs = template.FuncMap{}
