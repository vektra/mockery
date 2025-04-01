package template

import "github.com/vektra/mockery/v3/template_funcs"

// MockData is the data used to generate a mock for some interface.
type MockData struct {
	InterfaceName string
	MockName      string
	TypeParams    []TypeParamData
	Methods       []MethodData
	TemplateData  map[string]any
}

func (m MockData) TypeConstraintTest() string {
	if len(m.TypeParams) == 0 {
		return ""
	}
	s := "["
	for idx, param := range m.TypeParams {
		if idx != 0 {
			s += ", "
		}
		s += template_funcs.Exported(param.Name())
		s += " "
		s += param.TypeString()
	}
	s += "]"
	return s
}

func (m MockData) TypeConstraint() string {
	if len(m.TypeParams) == 0 {
		return ""
	}
	s := "["
	for idx, param := range m.TypeParams {
		if idx != 0 {
			s += ", "
		}
		s += template_funcs.Exported(param.Name())
		s += " "
		s += param.TypeString()
	}
	s += "]"
	return s
}

func (m MockData) TypeInstantiation() string {
	if len(m.TypeParams) == 0 {
		return ""
	}
	s := "["
	for idx, param := range m.TypeParams {
		if idx != 0 {
			s += ", "
		}
		s += template_funcs.Exported(param.Name())
	}
	s += "]"
	return s
}
