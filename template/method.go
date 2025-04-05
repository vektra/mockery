package template

import (
	"fmt"
	"strings"
)

// Method is the data which represents a method on some interface.
type Method struct {
	// Name is the method's name.
	Name string

	// Params represents all the arguments to the method.
	Params []Param

	// Returns represents all the return parameters of the method.
	Returns []Param

	// Scope represents the lexical scope of the method. Its primary function
	// is keeping track of all names visible in the current scope, which allows
	// the creation of new variables with guaranteed non-conflicting names.
	Scope *MethodScope
}

// ReturnStatement returns the string "return" if a method has return values.
// Otherwise, it returns an empty string.
func (m Method) ReturnStatement() string {
	if len(m.Returns) > 0 {
		return "return"
	}
	return ""
}

// Call returns a string containing the method call. This will usually need to be
// prefixed with a selector to specify which actual method to call. For example,
// if the method has a signature of "func Foo(s string) error", this method will
// return the string "Foo(s)". The name of each argument variable will be the same
// as what was generated post collision-resolution. Meaning, the argument variable
// name might be slightly altered from the original function if a naming collision
// was found.
func (m Method) Call() string {
	return fmt.Sprintf("%s(%s)", m.Name, m.ArgCallList())
}

// AcceptsContext returns whether or not the first argument of the method is a context.Context.
func (m Method) AcceptsContext() bool {
	if len(m.Params) > 0 && m.Params[0].TypeString() == "context.Context" {
		return true
	}
	return false
}

// Signature returns the string representation of the method's signature. For example,
// if a method was declared as "func (b Bar) Foo(s string) error", this method will
// return "(s string) error"
func (m Method) Signature() string {
	return fmt.Sprintf("(%s) (%s)", m.ArgList(), m.ReturnArgList())
}

// Declaration returns the method name followed by its signature. For
// example, if a method was declared as "func (b Bar) Foo(s string) error", this method
// will return "Foo(s string) error"
func (m Method) Declaration() string {
	return m.Name + m.Signature()
}

func (m Method) ReturnsError() bool {
	// Yes I know that by convention the last return value is the error,
	// but to be technically correct, we have to check all return values.
	for _, ret := range m.Returns {
		if ret.Var.TypeString() == "error" {
			return true
		}
	}
	return false
}

func (m Method) HasParams() bool {
	return len(m.Params) > 0
}

func (m Method) HasReturns() bool {
	return len(m.Returns) > 0
}

// ArgList is the string representation of method parameters, ex:
// 's string, n int, foo bar.Baz'.
func (m Method) ArgList() string {
	params := make([]string, len(m.Params))
	for i, p := range m.Params {
		params[i] = p.MethodArg()
	}
	return strings.Join(params, ", ")
}

// ArgTypeList returns the argument types in a comma-separated string, ex:
// `string, int, bar.Baz`
func (m Method) ArgTypeList() string {
	params := make([]string, len(m.Params))
	for i, p := range m.Params {
		params[i] = p.TypeString()
	}
	return strings.Join(params, ", ")
}

// ArgTypeListEllipsis returns the argument types in a comma-separated string, ex:
// `string, int, bar.Baz`. If the last argument is variadic, it will contain an
// ellipsis as would be expected in a variadic function definition.
func (m Method) ArgTypeListEllipsis() string {
	params := make([]string, len(m.Params))
	for i, p := range m.Params {
		params[i] = p.TypeStringEllipsis()
	}
	return strings.Join(params, ", ")
}

// ArgCallList is the string representation of method call parameters,
// ex: 's, n, foo'. In case of a last variadic parameter, it will be of
// the format 's, n, foos...'.
func (m Method) ArgCallList() string {
	return m.argCallListSlice(0, -1, true)
}

// ArgCallListNoEllipsis is the same as ArgCallList, except the last parameter, if
// variadic, will not contain an ellipsis.
func (m Method) ArgCallListNoEllipsis() string {
	return m.argCallListSlice(0, -1, false)
}

// argCallListSlice is similar to ArgCallList, but it allows specification of
// a slice range to use for the parameter lists. Specifying an integer less than
// 1 for end indicates to slice to the end of the parameters. As with regular
// Go slicing semantics, the end value is a non-inclusive index.
func (m Method) ArgCallListSlice(start, end int) string {
	return m.argCallListSlice(start, end, true)
}

func (m Method) ArgCallListSliceNoEllipsis(start, end int) string {
	return m.argCallListSlice(start, end, false)
}

func (m Method) argCallListSlice(start, end int, ellipsis bool) string {
	if end < 0 {
		end = len(m.Params)
	}
	if end == 1 && len(m.Params) == 0 {
		end = 0
	}
	paramsSlice := m.Params[start:end]
	params := make([]string, len(paramsSlice))
	for i, p := range paramsSlice {
		params[i] = p.CallName(ellipsis)
	}
	return strings.Join(params, ", ")
}

// ReturnArgTypeList is the string representation of method return
// types, ex: 'bar.Baz', '(string, error)'.
func (m Method) ReturnArgTypeList() string {
	params := make([]string, len(m.Returns))
	for i, p := range m.Returns {
		params[i] = p.TypeString()
	}
	if len(m.Returns) > 1 {
		return fmt.Sprintf("(%s)", strings.Join(params, ", "))
	}
	return strings.Join(params, ", ")
}

// ReturnArgNameList is the string representation of values being
// returned from the method, ex: 'foo', 's, err'.
func (m Method) ReturnArgNameList() string {
	params := make([]string, len(m.Returns))
	for i, p := range m.Returns {
		params[i] = p.Name()
	}
	return strings.Join(params, ", ")
}

// ReturnArgList returns the name and types of the return values. For example:
// "foo int, bar string, err error"
func (m Method) ReturnArgList() string {
	params := make([]string, len(m.Returns))
	for i, p := range m.Returns {
		params[i] = p.Name() + " " + p.TypeString()
	}
	return strings.Join(params, ", ")
}

func (m Method) IsVariadic() bool {
	return len(m.Params) > 0 && m.Params[len(m.Params)-1].Variadic
}
