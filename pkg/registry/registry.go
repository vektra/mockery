package registry

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"github.com/rs/zerolog"
	"golang.org/x/tools/go/packages"
)

// Registry encapsulates types information for the source and mock
// destination package. For the mock package, it tracks the list of
// imports and ensures there are no conflicts in the imported package
// qualifiers.
type Registry struct {
	dstPkgPath       string
	srcPkg           *packages.Package
	srcPkgName       string
	aliases          map[string]string
	imports          map[string]*Package
	importQualifiers map[string]*Package
}

// New loads the source package info and returns a new instance of
// Registry.
func New(srcPkg *packages.Package, dstPkgPath string) (*Registry, error) {
	return &Registry{
		dstPkgPath:       dstPkgPath,
		srcPkg:           srcPkg,
		srcPkgName:       srcPkg.Name,
		aliases:          parseImportsAliases(srcPkg.Syntax),
		imports:          make(map[string]*Package),
		importQualifiers: make(map[string]*Package),
	}, nil
}

func (r Registry) SrcPkg() *packages.Package {
	return r.srcPkg
}

// SrcPkgName returns the name of the source package.
func (r Registry) SrcPkgName() string {
	return r.srcPkg.Name
}

// LookupInterface returns the underlying interface definition of the
// given interface name.
func (r Registry) LookupInterface(name string) (*types.Interface, *types.TypeParamList, error) {
	obj := r.SrcPkg().Types.Scope().Lookup(name)
	if obj == nil {
		return nil, nil, fmt.Errorf("interface not found: %s", name)
	}

	if !types.IsInterface(obj.Type()) {
		return nil, nil, fmt.Errorf("%s (%s) is not an interface", name, obj.Type())
	}

	var tparams *types.TypeParamList
	named, ok := obj.Type().(*types.Named)
	if ok {
		tparams = named.TypeParams()
	}

	return obj.Type().Underlying().(*types.Interface).Complete(), tparams, nil
}

// MethodScope returns a new MethodScope.
func (r *Registry) MethodScope() *MethodScope {
	return NewMethodScope(r)
}

// AddImport adds the given package to the set of imports. It generates a
// suitable alias if there are any conflicts with previously imported
// packages.
func (r *Registry) AddImport(ctx context.Context, pkg *types.Package) *Package {
	log := zerolog.Ctx(ctx)
	path := pkg.Path()
	log.Debug().Str("method", "AddImport").Str("src-pkg-path", path).Str("dst-pkg-path", r.dstPkgPath).Msg("adding import")
	if path == r.dstPkgPath {
		return nil
	}

	if imprt, ok := r.imports[path]; ok {
		return imprt
	}

	imprt := Package{pkg: pkg, Alias: r.aliases[path]}
	var aliasSuggestion string
	for i := 0; ; i++ {
		if _, conflict := r.importQualifiers[aliasSuggestion]; conflict {
			aliasSuggestion = fmt.Sprintf("%s%d", imprt.Qualifier(), i)
			continue
		}
		imprt.Alias = aliasSuggestion
		break
	}

	r.imports[path] = &imprt
	r.importQualifiers[imprt.Qualifier()] = &imprt
	return &imprt
}

// Imports returns the list of imported packages. The list is sorted by
// path.
func (r Registry) Imports() []*Package {
	imports := make([]*Package, 0, len(r.imports))
	for _, imprt := range r.imports {
		imports = append(imports, imprt)
	}
	sort.Slice(imports, func(i, j int) bool {
		return imports[i].Path() < imports[j].Path()
	})
	return imports
}

func pkgInfoFromPath(srcDir string, mode packages.LoadMode) (*packages.Package, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: mode,
		Dir:  srcDir,
	})
	if err != nil {
		return nil, err
	}
	if len(pkgs) == 0 {
		return nil, errors.New("package not found")
	}
	if len(pkgs) > 1 {
		return nil, errors.New("found more than one package")
	}
	if errs := pkgs[0].Errors; len(errs) != 0 {
		if len(errs) == 1 {
			return nil, errs[0]
		}
		return nil, fmt.Errorf("%s (and %d more errors)", errs[0], len(errs)-1)
	}
	return pkgs[0], nil
}

func pkgInDir(pkgName, dir string) bool {
	currentPkg, err := pkgInfoFromPath(dir, packages.NeedName)
	if err != nil {
		return false
	}
	return currentPkg.Name == pkgName || currentPkg.Name+"_test" == pkgName
}

func parseImportsAliases(syntaxTree []*ast.File) map[string]string {
	aliases := make(map[string]string)
	for _, syntax := range syntaxTree {
		for _, imprt := range syntax.Imports {
			if imprt.Name != nil && imprt.Name.Name != "." && imprt.Name.Name != "_" {
				aliases[strings.Trim(imprt.Path.Value, `"`)] = imprt.Name.Name
			}
		}
	}
	return aliases
}
