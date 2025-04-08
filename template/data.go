package template

// Data is the template data used to render the mock template.
type Data struct {
	// PkgName is the name of the package chosen for the template.
	PkgName string
	// Registry chiefly maintains the list of imports that are required in the
	// rendered template file.
	Registry *Registry
	// SrcPkgQualifier is the qualifier used for the source package, if any.
	// For example, if the source package is different from the package the template
	// is rendered into, this string will contain something like "foo.", where
	// "foo" is the alias or package name of the source package.
	SrcPkgQualifier string
	// Interfaces is the list of interfaces being rendered in the template.
	Interfaces Interfaces
	// TemplateData is a schemaless map containing parameters from configuration
	// you may consume in your template.
	TemplateData TemplateData
}

func (d Data) Imports() Packages {
	return d.Registry.Imports()
}

func NewData(
	pkgName string,
	srcPkgQualifier string,
	imports Packages,
	interfaces Interfaces,
	templateData TemplateData,
	registry *Registry,
) Data {
	return Data{
		Interfaces:      interfaces,
		PkgName:         pkgName,
		Registry:        registry,
		TemplateData:    templateData,
		SrcPkgQualifier: srcPkgQualifier,
	}
}
