package mockery

import (
	"go/ast"
	"go/build"
	"go/importer"
	"go/types"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/tools/go/loader"
)

type Parser struct {
	configMapping    map[string][]*ast.File
	pathToInterfaces map[string][]string
	pathToASTFile    map[string]*ast.File
	parserPackages   []*types.Package
	conf             loader.Config
}

func NewParser() *Parser {
	var conf loader.Config

	conf.TypeCheckFuncBodies = func(_ string) bool { return false }
	conf.TypeChecker.DisableUnusedImportCheck = true
	conf.TypeChecker.Importer = importer.Default()

	// Initialize the build context (e.g. GOARCH/GOOS fields) so we can use it for respecting
	// build tags during Parse.
	buildCtx := build.Default
	conf.Build = &buildCtx

	return &Parser{
		parserPackages:   make([]*types.Package, 0),
		configMapping:    make(map[string][]*ast.File),
		pathToInterfaces: make(map[string][]string),
		pathToASTFile:    make(map[string]*ast.File),
		conf:             conf,
	}
}

func (p *Parser) AddBuildTags(buildTags ...string) {
	p.conf.Build.BuildTags = append(p.conf.Build.BuildTags, buildTags...)
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

		// If go/build would ignore this file, e.g. based on build tags, also ignore it here.
		//
		// (Further coupling with go internals and x/tools may of course bear a cost eventually
		// e.g. https://github.com/vektra/mockery/pull/117#issue-199337071, but should add
		// worthwhile consistency in this tool's behavior in the meantime.)
		match, matchErr := p.conf.Build.MatchFile(dir, fname)
		if matchErr != nil {
			return matchErr
		}
		if !match {
			continue
		}

		f, parseErr := p.conf.ParseFile(fpath, nil)
		if parseErr != nil {
			return parseErr
		}

		p.configMapping[path] = append(p.configMapping[path], f)
		p.pathToASTFile[fpath] = f
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
		for path, fi := range p.pathToASTFile {
			nv := NewNodeVisitor()
			ast.Walk(nv, fi)
			p.pathToInterfaces[path] = nv.DeclaredInterfaces()
		}
		wg.Done()
	}()

	// Type-check a package consisting of this file.
	// Type information for the imported packages
	// comes from $GOROOT/pkg/$GOOS_$GOOARCH/fmt.a.
	for path, files := range p.configMapping {
		p.conf.CreateFromFiles(path, files...)
	}

	prog, err := p.conf.Load()
	if err != nil {
		return err
	}

	for _, pkgInfo := range prog.Created {
		p.parserPackages = append(p.parserPackages, pkgInfo.Pkg)
	}

	wg.Wait()
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
	iFaces := p.pathToInterfaces[pkg.Path()]
	for i := 0; i < len(iFaces); i++ {
		iface := iFaces[i]
		if iface == name {
			list := make([]*Interface, 0)
			file := p.pathToASTFile[pkg.Path()]
			list = p.packageInterfaces(pkg, file, []string{name}, list)
			return list[0]
		}
	}

	return nil
}

type Interface struct {
	Name      string
	Path      string
	File      *ast.File
	Pkg       *types.Package
	Type      *types.Interface
	NamedType *types.Named
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
	for _, pkg := range p.parserPackages {
		path := pkg.Path()
		declaredIfaces := p.pathToInterfaces[path]
		astFile := p.pathToASTFile[path]
		ifaces = p.packageInterfaces(pkg, astFile, declaredIfaces, ifaces)
	}

	sort.Sort(ifaces)
	return ifaces
}

func (p *Parser) packageInterfaces(pkg *types.Package, file *ast.File, declaredInterfaces []string, ifaces []*Interface) []*Interface {
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
			Name:      name,
			Pkg:       pkg,
			Path:      pkg.Path(),
			Type:      iface.Complete(),
			NamedType: typ,
			File:      file,
		}

		ifaces = append(ifaces, elem)
	}

	return ifaces
}
