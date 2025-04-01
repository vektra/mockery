package template

// Data is the template data used to render the mock template.
type Data struct {
	PkgName         string
	SrcPkgQualifier string
	Imports         Packages
	Mocks           Mocks
	TemplateData    map[string]any
}
