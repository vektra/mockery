package pkg

import (
	"go/ast"

	"golang.org/x/tools/go/packages"
)

type Interface struct {
	Name     string // Name of the type to be mocked.
	FileName string
	File     *ast.File
	Pkg      *packages.Package
}
