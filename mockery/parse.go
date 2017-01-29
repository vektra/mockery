package mockery

import (
	"go/ast"
	"go/importer"
	"go/types"
	"io/ioutil"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/loader"
)

type Parser struct {
	file           *ast.File
	path           string
	nameToASTFiles map[string][]*ast.File
	parserPackages []*types.Package
}

func NewParser() *Parser {
	return &Parser{
		parserPackages: make([]*types.Package, 0),
		nameToASTFiles: make(map[string][]*ast.File),
	}
}

func (p *Parser) Parse(path string) error {
	dir := filepath.Dir(path)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	var conf loader.Config

	conf.TypeCheckFuncBodies = func(_ string) bool { return false }
	conf.TypeChecker.DisableUnusedImportCheck = true
	conf.TypeChecker.Importer = importer.Default()

	for _, fi := range files {
		if filepath.Ext(fi.Name()) != ".go" || strings.HasSuffix(fi.Name(), "_test.go") {
			continue
		}

		fpath := filepath.Join(dir, fi.Name())
		f, err := conf.ParseFile(fpath, nil)
		if err != nil {
			return err
		}

		if fi.Name() == filepath.Base(path) {
			p.file = f
		}
		p.nameToASTFiles[f.Name.String()] = append(p.nameToASTFiles[f.Name.String()], f)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	p.path = abs

	// Type-check a package consisting of this file.
	// Type information for the imported packages
	// comes from $GOROOT/pkg/$GOOS_$GOOARCH/fmt.a.
	for _, files := range p.nameToASTFiles {
		conf.CreateFromFiles(abs, files...)
	}

	prog, err := conf.Load()
	if err != nil {
		return err
	}

	for _, pkgInfo := range prog.Created {
		p.parserPackages = append(p.parserPackages, pkgInfo.Pkg)
	}

	return nil
}

func (p *Parser) Find(name string) (*Interface, error) {
	for _, pkg := range p.parserPackages {
		if iface := p.FindInPackage(name, pkg); iface != nil {
			return iface, nil
		}
	}
	return nil, ErrNotInterface
}

func (p *Parser) FindInPackage(name string, pkg *types.Package) *Interface {
	obj := pkg.Scope().Lookup(name)
	if obj == nil {
		return nil
	}

	typ := obj.Type().(*types.Named)

	name = typ.Obj().Name()

	iface := typ.Underlying().(*types.Interface).Complete()

	return &Interface{name, p.path, p.file, pkg, iface, typ}
}

/*
func (p *Parser) FindOld(name string) (*Interface, error) {
	for _, decl := range p.file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok {
			for _, spec := range gen.Specs {
				if typespec, ok := spec.(*ast.TypeSpec); ok {
					if typespec.Name.Name == name {
						if iface, ok := typespec.Type.(*ast.InterfaceType); ok {
							return &Interface{name, p.path, p.file, iface}, nil
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
*/

type Interface struct {
	Name      string
	Path      string
	File      *ast.File
	Pkg       *types.Package
	Type      *types.Interface
	NamedType *types.Named
}

func (p *Parser) getFileForInterfaceName(name string) *ast.File {
	for _, astFiles := range p.nameToASTFiles {
		for _, file := range astFiles {
			if file.Scope.Lookup(name) != nil {
				return file
			}
		}
	}
	return p.file
}

func (p *Parser) Interfaces() (ifaces []*Interface) {
	for _, pkg := range p.parserPackages {
		ifaces = p.packageInterfaces(pkg, ifaces)
	}
	return
}

func (p *Parser) packageInterfaces(pkg *types.Package, ifaces []*Interface) []*Interface {

	scope := pkg.Scope()

	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		typ, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}

		name = typ.Obj().Name()

		iface, ok := typ.Underlying().(*types.Interface)
		if !ok {
			continue
		}

		ifaces = append(
			ifaces,
			&Interface{
				name, p.path, p.getFileForInterfaceName(name), pkg,
				iface.Complete(), typ,
			},
		)
	}

	return ifaces
}
