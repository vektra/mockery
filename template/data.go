package template

// Data is the template data used to render the mock template.
type Data struct {
	PkgName         string
	SrcPkgQualifier string
	Imports         Packages
	Mocks           []MockData
	TemplateData    map[string]any
}

// MocksSomeMethod returns true of any one of the Mocks has at least 1
// method.
func (d Data) MocksSomeMethod() bool {
	for _, m := range d.Mocks {
		if len(m.Methods) > 0 {
			return true
		}
	}

	return false
}
