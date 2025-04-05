package template

import "github.com/vektra/mockery/v3/template_funcs"

// Interface is the data used to generate a mock for some interface.
type Interface struct {
	// Name is the name of the original interface.
	Name string
	// StructName is the chosen name for the struct that will implement the interface.
	StructName   string
	TypeParams   []TypeParam
	Methods      []Method
	TemplateData TemplateData
}

func (m Interface) TypeConstraintTest() string {
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

func (m Interface) TypeConstraint() string {
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

func (m Interface) TypeInstantiation() string {
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
