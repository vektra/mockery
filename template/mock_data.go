package template

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
		s += exported(param.Name())
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
		s += exported(param.Name())
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
		s += exported(param.Name())
	}
	s += "]"
	return s
}
