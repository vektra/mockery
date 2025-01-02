package pkg

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"github.com/vektra/mockery/v2/pkg/logging"
	"golang.org/x/tools/go/packages"
)

type fileEntry struct {
	fileName         string
	pkg              *packages.Package
	syntax           *ast.File
	interfaces       []string
	disableFuncMocks bool
}

func (f *fileEntry) ParseInterfaces(ctx context.Context) {
	nv := NewNodeVisitor(ctx, f.disableFuncMocks)
	ast.Walk(nv, f.syntax)
	f.interfaces = nv.DeclaredInterfaces()
}

type packageLoadEntry struct {
	pkgs []*packages.Package
	err  error
}

type Parser struct {
	files             []*fileEntry
	entriesByFileName map[string]*fileEntry
	parserPackages    []*types.Package
	conf              packages.Config
	packageLoadCache  map[string]packageLoadEntry
	disableFuncMocks  bool
}

func ParserDisableFuncMocks(disable bool) func(*Parser) {
	return func(p *Parser) {
		p.disableFuncMocks = disable
	}
}

func NewParser(buildTags []string, opts ...func(*Parser)) *Parser {
	var conf packages.Config
	conf.Mode = packages.NeedTypes |
		packages.NeedTypesSizes |
		packages.NeedSyntax |
		packages.NeedTypesInfo |
		packages.NeedImports |
		packages.NeedName |
		packages.NeedFiles |
		packages.NeedCompiledGoFiles

	if len(buildTags) > 0 {
		conf.BuildFlags = []string{"-tags", strings.Join(buildTags, ",")}
	}
	p := &Parser{
		parserPackages:    make([]*types.Package, 0),
		entriesByFileName: map[string]*fileEntry{},
		conf:              conf,
		packageLoadCache:  map[string]packageLoadEntry{},
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Parser) loadPackages(fpath string) ([]*packages.Package, error) {
	if result, ok := p.packageLoadCache[filepath.Dir(fpath)]; ok {
		return result.pkgs, result.err
	}
	pkgs, err := packages.Load(&p.conf, "file="+fpath)
	p.packageLoadCache[fpath] = packageLoadEntry{pkgs, err}
	return pkgs, err
}

func (p *Parser) ParsePackages(ctx context.Context, packageNames []string) error {
	log := zerolog.Ctx(ctx)

	packages, err := packages.Load(&p.conf, packageNames...)
	if err != nil {
		return err
	}
	for _, pkg := range packages {
		if len(pkg.GoFiles) == 0 {
			continue
		}
		for _, err := range pkg.Errors {
			log.Err(err).Msg("encountered error when loading package")
		}
		if len(pkg.Errors) != 0 {
			return errors.New("error occurred when loading packages")
		}
		for fileIdx, file := range pkg.GoFiles {
			log.Debug().
				Str("package", pkg.PkgPath).
				Str("file", file).
				Msgf("found file")
			entry := fileEntry{
				fileName:         file,
				pkg:              pkg,
				syntax:           pkg.Syntax[fileIdx],
				disableFuncMocks: p.disableFuncMocks,
			}
			entry.ParseInterfaces(ctx)
			p.files = append(p.files, &entry)
			p.entriesByFileName[file] = &entry
		}
	}
	return nil
}

// DEPRECATED: Parse is part of the deprecated, legacy mockery behavior. This is not
// used when the packages feature is enabled.
func (p *Parser) Parse(ctx context.Context, path string) error {
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

	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, fi := range files {
		log := zerolog.Ctx(ctx).With().
			Str(logging.LogKeyDir, dir).
			Str(logging.LogKeyFile, fi.Name()).
			Logger()
		ctx = log.WithContext(ctx)

		if filepath.Ext(fi.Name()) != ".go" || strings.HasSuffix(fi.Name(), "_test.go") || strings.HasPrefix(fi.Name(), "mock_") {
			continue
		}

		log.Debug().Msgf("parsing")

		fname := fi.Name()
		fpath := filepath.Join(dir, fname)
		if _, ok := p.entriesByFileName[fpath]; ok {
			continue
		}

		pkgs, err := p.loadPackages(fpath)
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
			entry := fileEntry{
				fileName: f,
				pkg:      pkg,
				syntax:   pkg.Syntax[idx],
			}
			p.files = append(p.files, &entry)
			p.entriesByFileName[f] = &entry
		}
	}

	return nil
}

func (p *Parser) Load(ctx context.Context) error {
	for _, entry := range p.files {
		entry.ParseInterfaces(ctx)
	}
	return nil
}

func (p *Parser) Find(name string) (*Interface, error) {
	for _, entry := range p.files {
		for _, iface := range entry.interfaces {
			if iface == name {
				list := p.packageInterfaces(entry.pkg.Types, entry.fileName, []string{name}, nil)
				if len(list) > 0 {
					return list[0], nil
				}
			}
		}
	}
	return nil, ErrNotInterface
}

func (p *Parser) Interfaces() []*Interface {
	ifaces := make(sortableIFaceList, 0)
	for _, entry := range p.files {
		declaredIfaces := entry.interfaces
		ifaces = p.packageInterfaces(entry.pkg.Types, entry.fileName, declaredIfaces, ifaces)
	}

	sort.Sort(ifaces)
	return ifaces
}

func (p *Parser) packageInterfaces(
	pkg *types.Package,
	fileName string,
	declaredInterfaces []string,
	ifaces []*Interface,
) []*Interface {
	scope := pkg.Scope()
	for _, name := range declaredInterfaces {
		obj := scope.Lookup(name)
		if obj == nil {
			continue
		}

		var typ *types.Named
		var name string

		ttyp := obj.Type()

		if talias, ok := obj.Type().(*types.Alias); ok {
			name = talias.Obj().Name()
			ttyp = types.Unalias(obj.Type())
		}

		typ, ok := ttyp.(*types.Named)
		if !ok {
			continue
		}

		if name == "" {
			name = typ.Obj().Name()
		}

		if typ.Obj().Pkg() == nil {
			continue
		}

		elem := &Interface{
			Name:          name,
			Pkg:           pkg,
			QualifiedName: pkg.Path(),
			FileName:      fileName,
			NamedType:     typ,
		}

		iface, ok := typ.Underlying().(*types.Interface)
		if ok {
			elem.IsFunction = false
			elem.ActualInterface = iface
		} else {
			sig, ok := typ.Underlying().(*types.Signature)
			if !ok {
				continue
			}
			elem.IsFunction = true
			elem.SingleFunction = &Method{Name: "Execute", Signature: sig}
		}

		ifaces = append(ifaces, elem)
	}

	return ifaces
}

type Method struct {
	Name      string
	Signature *types.Signature
}

type TypesPackage interface {
	Name() string
	Path() string
}

// Interface type represents the target type that we will generate a mock for.
// It could be an interface, or a function type.
// Function type emulates: an interface it has 1 method with the function signature
// and a general name, e.g. "Execute".
type Interface struct {
	Name            string // Name of the type to be mocked.
	QualifiedName   string // Path to the package of the target type.
	FileName        string
	File            *ast.File
	Pkg             TypesPackage
	NamedType       *types.Named
	IsFunction      bool             // If true, this instance represents a function, otherwise it's an interface.
	ActualInterface *types.Interface // Holds the actual interface type, in case it's an interface.
	SingleFunction  *Method          // Holds the function type information, in case it's a function type.
}

func (iface *Interface) Methods() []*Method {
	if iface.IsFunction {
		return []*Method{iface.SingleFunction}
	}
	methods := make([]*Method, iface.ActualInterface.NumMethods())
	for i := 0; i < iface.ActualInterface.NumMethods(); i++ {
		fn := iface.ActualInterface.Method(i)
		methods[i] = &Method{Name: fn.Name(), Signature: fn.Type().(*types.Signature)}
	}
	return methods
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

type NodeVisitor struct {
	declaredInterfaces []string
	disableFuncMocks   bool
	ctx                context.Context
}

func NewNodeVisitor(ctx context.Context, disableFuncMocks bool) *NodeVisitor {
	return &NodeVisitor{
		declaredInterfaces: make([]string, 0),
		disableFuncMocks:   disableFuncMocks,
		ctx:                ctx,
	}
}

func (nv *NodeVisitor) DeclaredInterfaces() []string {
	return nv.declaredInterfaces
}

func (nv *NodeVisitor) add(ctx context.Context, n *ast.TypeSpec) {
	log := zerolog.Ctx(ctx)
	log.Debug().
		Str("node-name", n.Name.Name).
		Str("node-type", fmt.Sprintf("%T", n.Type)).
		Msg("found node with acceptable type for mocking")
	nv.declaredInterfaces = append(nv.declaredInterfaces, n.Name.Name)
}

func (nv *NodeVisitor) Visit(node ast.Node) ast.Visitor {
	log := zerolog.Ctx(nv.ctx)

	switch n := node.(type) {
	case *ast.TypeSpec:
		log := log.With().
			Str("node-name", n.Name.Name).
			Str("node-type", fmt.Sprintf("%T", n.Type)).
			Logger()

		switch n.Type.(type) {
		case *ast.FuncType:
			if nv.disableFuncMocks {
				break
			}
			nv.add(nv.ctx, n)
		case *ast.InterfaceType, *ast.IndexExpr, *ast.IndexListExpr, *ast.SelectorExpr, *ast.Ident:
			nv.add(nv.ctx, n)
		default:
			log.Debug().Msg("found node with unacceptable type for mocking. Rejecting.")
		}
	}
	return nv
}
