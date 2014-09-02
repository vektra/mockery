package mockery

import (
	"go/ast"
	"go/parser"
	"go/token"
)

type Parser struct {
	file *ast.File
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

	p.file = f
	return nil
}

func (p *Parser) Find(name string) (ast.Expr, error) {
	for _, decl := range p.file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gen.Specs {
				if typespec, ok := spec.(*ast.TypeSpec); ok {
					if typespec.Name.Name == name {
						return typespec.Type, nil
					}
				}
			}
		}
	}
	return nil, nil
}
