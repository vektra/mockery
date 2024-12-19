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
	Config   *Config
}

func NewInterface(name string, filename string, file *ast.File, pkg *packages.Package, config *Config) *Interface {
	return &Interface{
		Name:     name,
		FileName: filename,
		File:     file,
		Pkg:      pkg,
		Config:   config,
	}
}
