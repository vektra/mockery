package mockery

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/types"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/tools/imports"

	"github.com/vektra/errors"
)

type Generator struct {
	buf bytes.Buffer

	ip    bool
	iface *Interface
}

func NewGenerator(iface *Interface) *Generator {
	return &Generator{
		iface: iface,
	}
}

func (g *Generator) GenerateIPPrologue() {
	g.ip = true

	g.printf("package %s\n\n", g.iface.File.Name)

	g.printf("import \"github.com/stretchr/testify/mock\"\n\n")
	if g.iface.File.Imports == nil {
		return
	}

	for _, imp := range g.iface.File.Imports {
		if imp.Name == nil {
			g.printf("import %s\n", imp.Path.Value)
		} else {
			g.printf("import %s %s\n", imp.Name.Name, imp.Path.Value)
		}
	}

	g.printf("\n")
}

func (g *Generator) mockName() string {
	if g.ip {
		if ast.IsExported(g.iface.Name) {
			return "Mock" + g.iface.Name
		} else {
			first := true
			return "mock" + strings.Map(func(r rune) rune {
				if first {
					first = false
					return unicode.ToUpper(r)
				}
				return r
			}, g.iface.Name)
		}
	}

	return g.iface.Name
}

func (g *Generator) GeneratePrologue(pkg string) {
	g.printf("package %v\n\n", pkg)

	local, err := fullPkgName(g.iface.Path)
	if err != nil {
		panic("unable to figure out path for package")
	}

	g.printf("import \"%s\"\n", local)

	g.printf("import \"github.com/stretchr/testify/mock\"\n\n")
	if g.iface.File.Imports == nil {
		return
	}

	for _, imp := range g.iface.File.Imports {
		if imp.Name == nil {
			g.printf("import %s\n", imp.Path.Value)
		} else {
			g.printf("import %s %s\n", imp.Name.Name, imp.Path.Value)
		}
	}

	g.printf("\n")
}

func (g *Generator) GeneratePrologueNote(note string) {
	if note != "" {
		g.printf("\n")
		for _, n := range strings.Split(note, "\\n") {
			g.printf("// %s\n", n)
		}
		g.printf("\n")
	}
}

var ErrNotInterface = errors.New("expression not an interface")

func (g *Generator) printf(s string, vals ...interface{}) {
	fmt.Fprintf(&g.buf, s, vals...)
}

var builtinTypes = map[string]bool{
	"ComplexType": true,
	"FloatType":   true,
	"IntegerType": true,
	"Type":        true,
	"Type1":       true,
	"bool":        true,
	"byte":        true,
	"complex128":  true,
	"complex64":   true,
	"error":       true,
	"float32":     true,
	"float64":     true,
	"int":         true,
	"int16":       true,
	"int32":       true,
	"int64":       true,
	"int8":        true,
	"rune":        true,
	"string":      true,
	"uint":        true,
	"uint16":      true,
	"uint32":      true,
	"uint64":      true,
	"uint8":       true,
	"uintptr":     true,
}

type namer interface {
	Name() string
}

func renderType(t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		o := t.Obj()
		if o.Pkg() == nil || o.Pkg().Name() == "main" {
			return o.Name()
		} else {
			return o.Pkg().Name() + "." + o.Name()
		}
	case *types.Basic:
		return t.Name()
	case *types.Pointer:
		return "*" + renderType(t.Elem())
	case *types.Slice:
		return "[]" + renderType(t.Elem())
	case *types.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), renderType(t.Elem()))
	case *types.Signature:
		switch t.Results().Len() {
		case 0:
			return fmt.Sprintf(
				"func(%s)",
				renderTypeTuple(t.Params()),
			)
		case 1:
			return fmt.Sprintf(
				"func(%s) %s",
				renderTypeTuple(t.Params()),
				renderType(t.Results().At(0).Type()),
			)
		default:
			return fmt.Sprintf(
				"func(%s)(%s)",
				renderTypeTuple(t.Params()),
				renderTypeTuple(t.Results()),
			)
		}
	case *types.Map:
		kt := renderType(t.Key())
		vt := renderType(t.Elem())

		return fmt.Sprintf("map[%s]%s", kt, vt)
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			return "chan " + renderType(t.Elem())
		case types.RecvOnly:
			return "<-chan " + renderType(t.Elem())
		default:
			return "chan<- " + renderType(t.Elem())
		}
	case *types.Struct:
		var fields []string

		for i := 0; i < t.NumFields(); i++ {
			f := t.Field(i)

			if f.Anonymous() {
				fields = append(fields, renderType(f.Type()))
			} else {
				fields = append(fields, fmt.Sprintf("%s %s", f.Name(), renderType(f.Type())))
			}
		}

		return fmt.Sprintf("struct{%s}", strings.Join(fields, ";"))
	case *types.Interface:
		if t.NumMethods() != 0 {
			panic("Unable to mock inline interfaces with methods")
		}

		return "interface{}"
	case namer:
		return t.Name()
	default:
		panic(fmt.Sprintf("un-namable type: %#v (%T)", t, t))
	}
}

func renderTypeTuple(tup *types.Tuple) string {
	var parts []string

	for i := 0; i < tup.Len(); i++ {
		v := tup.At(i)

		parts = append(parts, renderType(v.Type()))
	}

	return strings.Join(parts, " , ")
}

func isNillable(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Pointer, *types.Array, *types.Map, *types.Interface, *types.Signature, *types.Chan, *types.Slice:
		return true
	case *types.Named:
		return isNillable(t.Underlying())
	}
	return false
}

type paramList struct {
	Names   []string
	Types   []string
	Params  []string
	Nilable []bool
}

func genList(list *types.Tuple, varadic bool) *paramList {
	var params paramList

	if list == nil {
		return &params
	}

	for i := 0; i < list.Len(); i++ {
		v := list.At(i)

		ts := renderType(v.Type())

		if varadic && i == list.Len()-1 {
			t := v.Type()
			switch t := t.(type) {
			case *types.Slice:
				ts = "..." + renderType(t.Elem())
			default:
				panic("bad varadic type!")
			}
		}

		pname := v.Name()

		if pname == "" {
			pname = fmt.Sprintf("_a%d", i)
		}

		params.Names = append(params.Names, pname)
		params.Types = append(params.Types, ts)
		params.Params = append(params.Params, fmt.Sprintf("%s %s", pname, ts))
		params.Nilable = append(params.Nilable, isNillable(v.Type()))
	}

	return &params
}

var ErrNotSetup = errors.New("not setup")

func (g *Generator) Generate() error {
	if g.iface == nil {
		return ErrNotSetup
	}

	local, err := fullPkgName(g.iface.Path)
	if err != nil {
		return err
	}

	g.printf("// %s is an autogenerated mock type for %s.%s\n", g.mockName(), local, g.iface.Name)
	g.printf("type %s struct {\n\tmock.Mock\n}\n\n", g.mockName())

	for i := 0; i < g.iface.Type.NumMethods(); i++ {
		fn := g.iface.Type.Method(i)

		ftype := fn.Type().(*types.Signature)
		fname := fn.Name()

		params := genList(ftype.Params(), ftype.Variadic())
		returns := genList(ftype.Results(), false)

		g.printf("// %s provides a mock function with given fields: %s\n", fname, strings.Join(params.Names, ", "))
		g.printf("func (_m *%s) %s(%s) ", g.mockName(), fname, strings.Join(params.Params, ", "))

		switch len(returns.Types) {
		case 0:
			g.printf("{\n")
		case 1:
			g.printf("%s {\n", returns.Types[0])
		default:
			g.printf("(%s) {\n", strings.Join(returns.Types, ", "))
		}

		formatParamNames := func() string {
			names := ""
			for i, name := range params.Names {
				if i > 0 {
					names += ", "
				}

				paramType := params.Types[i]
				// for variable args, move the ... to the end.
				if strings.Index(paramType, "...") == 0 {
					name += "..."
				}
				names += name
			}
			return names
		}

		if len(returns.Types) > 0 {
			g.printf("\tret := _m.Called(%s)\n\n", strings.Join(params.Names, ", "))

			var (
				ret []string
			)

			for idx, typ := range returns.Types {
				g.printf("\tvar r%d %s\n", idx, typ)
				g.printf("\tif rf, ok := ret.Get(%d).(func(%s) %s); ok {\n",
					idx, strings.Join(params.Types, ", "), typ)
				g.printf("\t\tr%d = rf(%s)\n", idx, formatParamNames())
				g.printf("\t} else {\n")
				if typ == "error" {
					g.printf("\t\tr%d = ret.Error(%d)\n", idx, idx)
				} else if returns.Nilable[idx] {
					g.printf("\t\tif ret.Get(%d) != nil {\n", idx)
					g.printf("\t\t\tr%d = ret.Get(%d).(%s)\n", idx, idx, typ)
					g.printf("\t\t}\n")
				} else {
					g.printf("\t\tr%d = ret.Get(%d).(%s)\n", idx, idx, typ)
				}
				g.printf("\t}\n\n")

				ret = append(ret, fmt.Sprintf("r%d", idx))
			}

			g.printf("\treturn %s\n", strings.Join(ret, ", "))
		} else {
			g.printf("\t_m.Called(%s)\n", strings.Join(params.Names, ", "))
		}

		g.printf("}\n")
	}

	return nil
}

func (g *Generator) Write(w io.Writer) error {
	opt := &imports.Options{Comments: true}
	res, err := imports.Process("mock.go", g.buf.Bytes(), opt)
	if err != nil {
		return err
	}

	w.Write(res)
	return nil
}

func fullPkgName(path string) (string, error) {
	goPath := os.Getenv("GOPATH")
	return filepath.Rel(filepath.Join(goPath, "src"), filepath.Dir(path))
}
