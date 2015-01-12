package mockery

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"io"
	"os"
	"path/filepath"
	"strings"

	"code.google.com/p/go.tools/imports"

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
		return "Mock" + g.iface.Name
	}

	return g.iface.Name
}

func (g *Generator) GeneratePrologue() {
	g.printf("package mocks\n\n")

	goPath := os.Getenv("GOPATH")

	local, err := filepath.Rel(filepath.Join(goPath, "src"), filepath.Dir(g.iface.Path))
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

var ErrNotInterface = errors.New("expression not an interface")

func (g *Generator) printf(s string, vals ...interface{}) {
	fmt.Fprintf(&g.buf, s, vals...)
}

func (g *Generator) typeString(typ ast.Expr) string {
	var buf bytes.Buffer

	err := printer.Fprint(&buf, g.iface.FSet, typ)
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func (g *Generator) genList(list *ast.FieldList, addNames bool) ([]string, []string, []string) {
	var (
		params []string
		names  []string
		types  []string
	)

	if list == nil {
		return params, names, types
	}

	for idx, param := range list.List {
		ts := g.typeString(param.Type)

		var pname string

		if addNames {
			if len(param.Names) == 0 {
				pname = fmt.Sprintf("_a%d", idx)
				names = append(names, pname)
				types = append(types, ts)
				params = append(params, fmt.Sprintf("%s %s", pname, ts))

				continue
			}

			for _, name := range param.Names {
				pname = name.Name
				names = append(names, pname)
				types = append(types, ts)
				params = append(params, fmt.Sprintf("%s %s", pname, ts))
			}
		} else {
			names = append(names, "")
			types = append(types, ts)
			params = append(params, ts)
		}
	}

	return names, types, params
}

var ErrNotSetup = errors.New("not setup")

func (g *Generator) Generate() error {
	if g.iface == nil {
		return ErrNotSetup
	}

	g.printf("type %s struct {\n\tmock.Mock\n}\n\n", g.mockName())

	for _, method := range g.iface.Type.Methods.List {
		ftype, ok := method.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		fname := method.Names[0].Name

		names, _, params := g.genList(ftype.Params, true)
		_, types, returs := g.genList(ftype.Results, false)

		g.printf("func (m *%s) %s(%s) ", g.mockName(), fname, strings.Join(params, ", "))

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
