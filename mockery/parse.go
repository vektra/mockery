package mockery

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type Parser struct {
	file *ast.File
	path string
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(path string) error {
	fset := token.NewFileSet()

	// Parse the file containing this very example
	// but stop after processing the imports.
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return err
	}

	p.path = path
	p.file = f
	return nil
}

func (p *Parser) Find(name string) (*Interface, error) {
	for _, decl := range p.file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gen.Specs {
				if typespec, ok := spec.(*ast.TypeSpec); ok {
					if typespec.Name.Name == name {
						if iface, ok := typespec.Type.(*ast.InterfaceType); ok {
							return &Interface{name, p.file, iface}, nil
						} else {
							return nil, ErrNotInterface
						}
					}
				}
			}
		}
	}
	return nil, nil
}

type Interface struct {
	Name string
	File *ast.File
	Type *ast.InterfaceType
}

func (p *Parser) Interfaces() []*Interface {
	var ifaces []*Interface

	for _, decl := range p.file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gen.Specs {
				if typespec, ok := spec.(*ast.TypeSpec); ok {
					if iface, ok := typespec.Type.(*ast.InterfaceType); ok {
						ifaces = append(ifaces, &Interface{typespec.Name.Name, p.file, iface})
					}
				}
			}
		}
	}

	return ifaces
}
