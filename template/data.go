package template

// Data is the template data used to render the mock template.
type Data struct {
	// PkgName is the name of the package chosen for the template.
	PkgName string
	// SrcPkgQualifier is the qualifier used for the source package, if any.
	// For example, if the source package is different from the package the template
	// is rendered into, this string will contain something like "foo.", where
	// "foo" is the alias or package name of the source package.
	SrcPkgQualifier string
	// Imports is the list of imports necessary for this template.
	Imports Packages
	// Interfaces is the list of interfaces being rendered in the template.
	Interfaces Interfaces
	// TemplateData is a schemaless map containing parameters from configuration
	// you may consume in your template.
	TemplateData map[string]any
}
