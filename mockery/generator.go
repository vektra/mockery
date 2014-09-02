package mockery

import (
	"bytes"
	"fmt"
	"go/ast"
	"io"
	"strings"

	"code.google.com/p/go.tools/imports"

	"github.com/vektra/errors"
)

type Generator struct {
	buf   bytes.Buffer
	iface *Interface
}

func NewGenerator(iface *Interface) *Generator {
	return &Generator{
		iface: iface,
	}
}

func (g *Generator) GeneratePrologue() {
	g.printf("package mocks\n\n")
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

var ErrNotInterface = errors.New("expression not an interface")

func (g *Generator) printf(s string, vals ...interface{}) {
	fmt.Fprintf(&g.buf, s, vals...)
}

func (g *Generator) typeString(typ ast.Expr) string {
	switch specific := typ.(type) {
	case *ast.Ident:
		return specific.Name
	case *ast.StarExpr:
		return "*" + g.typeString(specific.X)
	case *ast.ArrayType:
		if specific.Len == nil {
			return "[]" + g.typeString(specific.Elt)
		} else {
			var l string

			switch ls := specific.Len.(type) {
			case *ast.BasicLit:
				l = ls.Value
			default:
				panic(fmt.Sprintf("unable to figure out array length: %#v", specific.Len))
			}
			return "[" + l + "]" + g.typeString(specific.Elt)
		}
	case *ast.SelectorExpr:
		return g.typeString(specific.X) + "." + specific.Sel.Name
	case *ast.InterfaceType:
		if len(specific.Methods.List) == 0 {
			return "interface{}"
		} else {
			panic(fmt.Sprintf("unable to handle this interface type: %#v", specific))
		}
	case *ast.MapType:
		return "map[" + g.typeString(specific.Key) + "]" + g.typeString(specific.Value)
	case *ast.Ellipsis:
		return "..." + g.typeString(specific.Elt)
	default:
		panic(fmt.Sprintf("unable to handle type: %#v", typ))
	}
}

func (g *Generator) genList(list *ast.FieldList) ([]string, []string, []string) {
	var (
		params []string
		names  []string
		types  []string
	)

	if list == nil {
		return params, names, types
	}

	for _, param := range list.List {
		ts := g.typeString(param.Type)

		if len(param.Names) == 0 {
			names = append(names, "")
			types = append(types, ts)
			params = append(params, ts)
		} else {
			pname := param.Names[0].Name

			names = append(names, pname)
			types = append(types, ts)
			params = append(params, fmt.Sprintf("%s %s", pname, ts))
		}
	}

	return names, types, params
}

var ErrNotSetup = errors.New("not setup")

func (g *Generator) Generate() error {
	if g.iface == nil {
		return ErrNotSetup
	}

	g.printf("type %s struct {\n\tmock.Mock\n}\n\n", g.iface.Name)

	for _, method := range g.iface.Type.Methods.List {
		ftype, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		fname := method.Names[0].Name

		names, _, params := g.genList(ftype.Params)
		_, types, returs := g.genList(ftype.Results)

		g.printf("func (m *%s) %s(%s) ", g.iface.Name, fname, strings.Join(params, ", "))

		switch len(returs) {
		case 0:
			g.printf("{\n")
		case 1:
			g.printf("%s {\n", returs[0])
		default:
			g.printf("(%s) {\n", strings.Join(returs, ", "))
		}

		if len(types) > 0 {
			g.printf("\tret := m.Called(%s)\n\n", strings.Join(names, ", "))

			var ret []string

			for idx, typ := range types {
				if typ == "error" {
					g.printf("\tr%d := ret.Error(%d)\n", idx, idx)
				} else {
					g.printf("\tr%d := ret.Get(%d).(%s)\n", idx, idx, typ)
				}
				ret = append(ret, fmt.Sprintf("r%d", idx))
			}

			g.printf("\n\treturn %s\n", strings.Join(ret, ", "))

		} else {
			g.printf("\tm.Called(%s)\n", strings.Join(names, ", "))
		}

		g.printf("}\n")
	}

	return nil
}

func (g *Generator) Write(w io.Writer) error {
	opt := &imports.Options{}
	res, err := imports.Process("mock.go", g.buf.Bytes(), opt)
	if err != nil {
		return err
	}

	w.Write(res)
	return nil
}
