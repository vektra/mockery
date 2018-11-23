package mockery

import (
	"fmt"
	"go/ast"
	"go/types"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/packages"
)

type parserEntry struct {
	fileName   string
	pkg        *packages.Package
	syntax     *ast.File
	interfaces []string
}

type Parser struct {
	entries           []*parserEntry
	entriesByFileName map[string]*parserEntry
	packages          []*packages.Package
	parserPackages    []*types.Package
	conf              packages.Config
}

func NewParser(buildTags []string) *Parser {
	var conf packages.Config
	conf.Mode = packages.LoadSyntax
	if len(buildTags) > 0 {
		conf.BuildFlags = []string{"-tags", strings.Join(buildTags, ",")}
	}
	return &Parser{
		parserPackages:    make([]*types.Package, 0),
		entriesByFileName: map[string]*parserEntry{},
		conf:              conf,
	}
}

func (p *Parser) Parse(path string) error {
	// To support relative paths to mock targets w/ vendor deps, we need to provide eventual
	// calls to build.Context.Import with an absolute path. It needs to be absolute because
	// Import will only find the vendor directory if our target path for parsing is under
	// a "root" (GOROOT or a GOPATH). Only absolute paths will pass the prefix-based validation.
	//
	// For example, if our parse target is "./ifaces", Import will check if any "roots" are a
	// prefix of "ifaces" and decide to skip the vendor search.
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fi := range files {
		if filepath.Ext(fi.Name()) != ".go" || strings.HasSuffix(fi.Name(), "_test.go") {
			continue
		}

		fname := fi.Name()
		fpath := filepath.Join(dir, fname)
		if _, ok := p.entriesByFileName[fpath]; ok {
			continue
		}

		pkgs, err := packages.Load(&p.conf, "file="+fpath)
		if err != nil {
			return err
		}
		if len(pkgs) == 0 {
			continue
		}
		if len(pkgs) > 1 {
			names := make([]string, len(pkgs))
			for i, p := range pkgs {
				names[i] = p.Name
			}
			panic(fmt.Sprintf("file %s resolves to multiple packages: %s", fpath, strings.Join(names, ", ")))
		}

		pkg := pkgs[0]
		if len(pkg.Errors) > 0 {
			return pkg.Errors[0]
		}
		if len(pkg.GoFiles) == 0 {
			continue
		}

		for idx, f := range pkg.GoFiles {
			if _, ok := p.entriesByFileName[f]; ok {
				continue
			}
			entry := parserEntry{
				fileName: f,
				pkg:      pkg,
				syntax:   pkg.Syntax[idx],
			}
			p.entries = append(p.entries, &entry)
			p.entriesByFileName[f] = &entry
		}
		p.packages = append(p.packages, pkg)
	}

	return nil
}

type NodeVisitor struct {
	declaredInterfaces []string
}

func NewNodeVisitor() *NodeVisitor {
	return &NodeVisitor{
		declaredInterfaces: make([]string, 0),
	}
}

func (n *NodeVisitor) DeclaredInterfaces() []string {
	return n.declaredInterfaces
}

func (nv *NodeVisitor) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.TypeSpec:
		if _, ok := n.Type.(*ast.InterfaceType); ok {
			nv.declaredInterfaces = append(nv.declaredInterfaces, n.Name.Name)
		}
	}
	return nv
}

func (p *Parser) Load() error {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for _, entry := range p.entries {
			nv := NewNodeVisitor()
			ast.Walk(nv, entry.syntax)
			entry.interfaces = nv.DeclaredInterfaces()
		}
		wg.Done()
	}()
	wg.Wait()
	return nil
}

func (p *Parser) Find(name string) (*Interface, error) {
	for _, entry := range p.entries {
		for _, iface := range entry.interfaces {
			if iface == name {
				list := p.packageInterfaces(entry.pkg.Types, entry.syntax, entry.fileName, []string{name}, nil)
				if len(list) > 0 {
					return list[0], nil
				}
			}
		}
	}
	return nil, ErrNotInterface
}

type Interface struct {
	Name          string
	QualifiedName string
	FileName      string
	File          *ast.File
	Pkg           *types.Package
	Type          *types.Interface
	NamedType     *types.Named
}

type sortableIFaceList []*Interface

func (s sortableIFaceList) Len() int {
	return len(s)
}

func (s sortableIFaceList) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortableIFaceList) Less(i, j int) bool {
	return strings.Compare(s[i].Name, s[j].Name) == -1
}

func (p *Parser) Interfaces() []*Interface {
	ifaces := make(sortableIFaceList, 0)
	for _, entry := range p.entries {
		declaredIfaces := entry.interfaces
		astFile := entry.syntax
		ifaces = p.packageInterfaces(entry.pkg.Types, astFile, entry.fileName, declaredIfaces, ifaces)
	}

	sort.Sort(ifaces)
	return ifaces
}

func (p *Parser) packageInterfaces(
	pkg *types.Package,
	file *ast.File,
	fileName string,
	declaredInterfaces []string,
	ifaces []*Interface) []*Interface {
	scope := pkg.Scope()
	for _, name := range declaredInterfaces {
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

		if typ.Obj().Pkg() == nil {
			continue
		}

		elem := &Interface{
			Name:          name,
			Pkg:           pkg,
			QualifiedName: pkg.Path(),
			FileName:      fileName,
			Type:          iface.Complete(),
			NamedType:     typ,
			File:          file,
		}

		ifaces = append(ifaces, elem)
	}

	return ifaces
}
