package template

import (
	"go/types"
	"strings"
)

// Var represents a method variable/parameter.
//
// It should be created using a method scope instance.
type Var struct {
	vr *types.Var
	// typ is stored separately from `vr.Type()` because it's possible
	// for a variable to be replaced with another variable via replace-type.
	// In such a case, `vr.Type()` refers to the original type and `typ` refers
	// to the replacer type.
	typ     types.Type
	imports map[string]*Package
	pkgPath string

	Name string
}

func (v Var) Type() types.Type {
	return v.typ
}

// IsSlice returns whether the type (or the underlying type) is a slice.
func (v Var) IsSlice() bool {
	_, ok := v.Type().Underlying().(*types.Slice)
	return ok
}

// TypeString returns the variable type with the package qualifier in the
// format 'pkg.Type'.
func (v Var) TypeString() string {
	return types.TypeString(v.Type(), v.packageQualifier)
}

// packageQualifier is a types.Qualifier.
func (v Var) packageQualifier(pkg *types.Package) string {
	path := pkg.Path()
	if v.pkgPath != "" && v.pkgPath == path {
		return ""
	}
	return v.imports[path].Qualifier()
}

func nillable(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Pointer, *types.Array, *types.Map, *types.Interface, *types.Signature, *types.Chan, *types.Slice:
		return true
	case *types.Named, *types.Alias, *types.TypeParam:
		return nillable(t.Underlying())
	}
	return false
}

func (v Var) Nillable() bool {
	return nillable(v.Type())
}

func varName(vr *types.Var, suffix string) string {
	name := vr.Name()
	if name != "" && name != "_" {
		return name + suffix
	}

	name = varNameForType(vr.Type()) + suffix

	switch name {
	case "mock", "callInfo", "break", "default", "func", "interface", "select", "case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch", "const", "fallthrough", "if", "range", "type", "continue", "for",
		"import", "return", "var",
		// avoid shadowing basic types
		"string", "bool", "byte", "rune", "uintptr",
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "complex64", "complex128":
		name += "Param"
	}

	return name
}

// varNameForType generates a name for the variable using the type
// information.
//
// Examples:
// - string -> s
// - int -> n
// - chan int -> intCh
// - []a.MyType -> myTypes
// - map[string]int -> stringToInt
// - error -> err
// - a.MyType -> myType
func varNameForType(t types.Type) string {
	nestedType := func(t types.Type) string {
		if t, ok := t.(*types.Basic); ok {
			return deCapitalise(t.String())
		}
		return varNameForType(t)
	}

	switch t := t.(type) {
	case *types.Named:
		if t.Obj().Name() == "error" {
			return "err"
		}

		name := deCapitalise(t.Obj().Name())
		if name == t.Obj().Name() {
			name += "MoqParam"
		}

		return name

	case *types.Basic:
		return basicTypeVarName(t)

	case *types.Array:
		return nestedType(t.Elem()) + "s"

	case *types.Slice:
		return nestedType(t.Elem()) + "s"

	case *types.Struct: // anonymous struct
		return "val"

	case *types.Pointer:
		return varNameForType(t.Elem())

	case *types.Signature:
		return "fn"

	case *types.Interface: // anonymous interface
		return "ifaceVal"

	case *types.Map:
		return nestedType(t.Key()) + "To" + capitalise(nestedType(t.Elem()))

	case *types.Chan:
		return nestedType(t.Elem()) + "Ch"
	}

	return "v"
}

func basicTypeVarName(b *types.Basic) string {
	switch b.Info() {
	case types.IsBoolean:
		return "b"

	case types.IsInteger:
		return "n"

	case types.IsFloat:
		return "f"

	case types.IsString:
		return "s"
	}

	return "v"
}

func capitalise(s string) string   { return strings.ToUpper(s[:1]) + s[1:] }
func deCapitalise(s string) string { return strings.ToLower(s[:1]) + s[1:] }
