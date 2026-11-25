package template

// Interface is the data used to generate a mock for some interface.
type Interface struct {
	Comments Comments
	// ConstructorName is the configured name for the mock's constructor function.
	ConstructorName string
	Methods         []Method
	// Name is the name of the original interface.
	Name string
	// StructName is the chosen name for the struct that will implement the interface.
	StructName   string
	TemplateData TemplateData
	TypeParams   []TypeParam
}

func NewInterface(
	name string,
	structName string,
	typeParams []TypeParam,
	methods []Method,
	templateData TemplateData,
	comments Comments,
) Interface {
	return Interface{
		Name:         name,
		StructName:   structName,
		TypeParams:   typeParams,
		Methods:      methods,
		TemplateData: templateData,
		Comments:     comments,
	}
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
		s += param.Name()
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
		s += param.Name()
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
		s += param.Name()
	}
	s += "]"
	return s
}
