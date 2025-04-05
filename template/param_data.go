package template

import (
	"fmt"
	"strings"
)

// Param is the data which represents a parameter to some method of
// an interface.
type Param struct {
	Var      *Var
	Variadic bool
}

// Name returns the name of the parameter.
func (p Param) Name() string {
	return p.Var.Name
}

// MethodArg is the representation of the parameter in the function
// signature, ex: 'name a.Type'.
func (p Param) MethodArg() string {
	if p.Variadic {
		return fmt.Sprintf("%s ...%s", p.Name(), p.TypeString()[2:])
	}
	return fmt.Sprintf("%s %s", p.Name(), p.TypeString())
}

// CallName returns the string representation of the parameter to be
// used for a method call. For a variadic paramter, it will be of the
// format 'foos...' if ellipsis is true.
func (p Param) CallName(ellipsis bool) string {
	if ellipsis && p.Variadic {
		return p.Name() + "..."
	}
	return p.Name()
}

// TypeString returns the string representation of the type of the
// parameter.
func (p Param) TypeString() string {
	return p.Var.TypeString()
}

// TypeStringEllipsis returns the string representation of the type of the
// parameter. If it is a variadic parameter, it will be represented as a
// variadic parameter instead of a slice. For example instead of `[]string`,
// it will return `...string`.
func (p Param) TypeStringEllipsis() string {
	typeString := p.TypeString()
	if !p.Variadic {
		return typeString
	}
	return strings.Replace(typeString, "[]", "...", 1)
}

// TypeStringVariadicUnderlying returns the underlying type of a variadic parameter. For
// instance, if a function has a parameter defined as `foo ...int`, this function
// will return "int". If the parameter is not variadic, this will behave the same
// as `TypeString`.
func (p Param) TypeStringVariadicUnderlying() string {
	typeString := p.TypeStringEllipsis()
	return strings.Replace(typeString, "...", "", 1)
}
