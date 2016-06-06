package mockery

import (
	"go/ast"
	"go/importer"
	"go/types"
	"io/ioutil"
	"path/filepath"

	"golang.org/x/tools/go/loader"
)

type Parser struct {
	file *ast.File
	path string

	pkg *types.Package
}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(path string) error {
	dir := filepath.Dir(path)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	var astFiles []*ast.File
	var conf loader.Config

	conf.TypeCheckFuncBodies = func(_ string) bool { return false }
	conf.TypeChecker.DisableUnusedImportCheck = true
	conf.TypeChecker.Importer = importer.Default()

	for _, fi := range files {
		if filepath.Ext(fi.Name()) != ".go" {
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

		astFiles = append(astFiles, f)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	// Type-check a package consisting of this file.
	// Type information for the imported packages
	// comes from $GOROOT/pkg/$GOOS_$GOOARCH/fmt.a.
	conf.CreateFromFiles(abs, astFiles...)

	prog, err := conf.Load()
	if err != nil {
		return err
	} else if len(prog.Created) != 1 {
		panic("expected only one Created package")
	}

	p.path = abs
	p.pkg = prog.Created[0].Pkg

	return nil
}

func (p *Parser) Find(name string) (*Interface, error) {
	obj := p.pkg.Scope().Lookup(name)
	if obj == nil {
		return nil, ErrNotInterface
	}

	typ := obj.Type().(*types.Named)

	name = typ.Obj().Name()

	iface := typ.Underlying().(*types.Interface).Complete()

	return &Interface{name, p.path, p.file, p.pkg, iface}, nil
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
	Name string
	Path string
	File *ast.File
	Pkg  *types.Package
	Type *types.Interface
}

func (p *Parser) Interfaces() []*Interface {
	var ifaces []*Interface

	scope := p.pkg.Scope()

	for _, name := range scope.Names() {
		obj := p.pkg.Scope().Lookup(name)
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

		ifaces = append(ifaces, &Interface{name, p.path, p.file, p.pkg, iface.Complete()})
	}

	return ifaces
}
