// Package template provides data and functionality for rendering templates using mockery.
// The data and methods herein are guaranteed to be backwards compatible, as they
// are used directly by user-defined Go templates. It is safe to import or use
// this package for any purpose.
package template

import (
	"io"
	"text/template"

	"github.com/vektra/mockery/v3/template_funcs"
)

// Template represents the template requested for rendering.
type Template struct {
	tmpl *template.Template
}

// New returns a new instance of Template.
func New(templateString string, name string) (Template, error) {
	tmpl, err := template.New(name).Funcs(template_funcs.FuncMap).Parse(templateString)
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
