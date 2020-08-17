package pkg

import (
	"context"
	"go/ast"
	"go/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
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
	parserPackages    []*types.Package
	conf              *packages.Config
}

func NewParser(buildTags []string) *Parser {
	conf := &packages.Config{
		Mode:  packages.NeedFiles | packages.NeedImports | packages.NeedName | packages.NeedSyntax | packages.NeedTypes,
		Tests: false,
	}
	if len(buildTags) > 0 {
		conf.BuildFlags = []string{"-tags", strings.Join(buildTags, ",")}
	}
	return &Parser{
		parserPackages:    make([]*types.Package, 0),
		entriesByFileName: map[string]*parserEntry{},
		conf:              conf,
	}
}

func (p *Parser) Parse(ctx context.Context, pattern string) error {
	log := zerolog.Ctx(ctx)

	info, err := os.Stat(pattern)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if info != nil && (strings.HasPrefix(info.Name(), ".") || strings.HasPrefix(info.Name(), "_")) {
		log.Debug().Msgf("Not loading path %q as it is prefixed with either '.' or '_'.", pattern)
		return nil
	}

	var query string
	switch {
	case os.IsNotExist(err):
		// The pattern represents one or more packages and should be passed directly to the package loader.
		log.Debug().Msgf("Loading packages corresponding to pattern %q.", pattern)
		query = pattern

	case !info.IsDir():
		// A file should be passed directly to the package loader as a 'file' query.
		if filepath.Ext(pattern) != ".go" {
			return errors.Errorf("specified file %q cannot be parsed as it is not a source file", pattern)
		}

		pattern, err = filepath.Abs(pattern)
		if err != nil {
			return err
		}

		log.Debug().Msgf("Loading file %q.", pattern)
		query = "file=" + pattern

	case info.IsDir():
		// A directory must have its files parsed individually as the package loader does not accept directory queries.
		var dir []os.FileInfo
		dir, err = ioutil.ReadDir(pattern)
		if err != nil {
			return err
		}
		log.Debug().Msgf("Loading files in directory %q.", pattern)
		for _, fi := range dir {
			if fi.IsDir() || filepath.Ext(fi.Name()) != ".go" {
				continue
			}
			if err = p.Parse(ctx, filepath.Join(pattern, fi.Name())); err != nil {
				return err
			}
		}

	default:
		// This is theoretically impossible to reach due to the disjunction of cases operated above.
		return errors.Errorf("encountered unexpected situation when retrieving information about %q", pattern)
	}

	log.Debug().Msgf("parsing")

	pkgs, err := packages.Load(p.conf, query)
	if err != nil {
		return err
	} else if filepath.Ext(pattern) == ".go" && len(pkgs) > 1 {
		err := errors.Errorf("file %q maps to multiple packages (%d) instead of a single one", pattern, len(pkgs))
		log.Err(err).Msgf("invalid file content")
		return err
	}

	for _, pkg := range pkgs {
		log.Debug().Msgf("Parsed sources from %q.", query)
		if len(pkg.Errors) > 0 {
			return pkg.Errors[0]
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
		switch n.Type.(type) {
		case *ast.InterfaceType, *ast.FuncType:
			nv.declaredInterfaces = append(nv.declaredInterfaces, n.Name.Name)
		}
	}
	return nv
}

func (p *Parser) Load() error {
	for _, entry := range p.entries {
		nv := NewNodeVisitor()
		ast.Walk(nv, entry.syntax)
		entry.interfaces = nv.DeclaredInterfaces()
	}
	return nil
}

func (p *Parser) Find(name string) (*Interface, error) {
	for _, entry := range p.entries {
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

type Method struct {
	Name      string
	Signature *types.Signature
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
	Pkg             *types.Package
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

func (p *Parser) Interfaces() []*Interface {
	ifaces := make(sortableIFaceList, 0)
	for _, entry := range p.entries {
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
