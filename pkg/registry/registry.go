package registry

import (
	"context"
	"fmt"
	"go/types"
	"sort"

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
	path := pkg.Path()
	log := zerolog.Ctx(ctx).With().
		Str("method", "AddImport").
		Str("src-pkg-path", path).
		Str("dst-pkg-path", r.dstPkgPath).
		Logger()
	log.Debug().Msg("adding import")
	if path == r.dstPkgPath {
		log.Debug().Msg("path equals dst-pkg-path, not adding import")
		return nil
	} else {
		log.Debug().Msg("path does not equal dst-pkg-path, adding import")
	}

	if imprt, ok := r.imports[path]; ok {
		return imprt
	}

	imprt := Package{pkg: pkg}
	originalQualifier := imprt.Qualifier()
	var aliasSuggestion string = imprt.Qualifier()
	for i := 0; ; i++ {
		if _, conflict := r.importQualifiers[aliasSuggestion]; conflict {
			aliasSuggestion = fmt.Sprintf("%s%d", imprt.Qualifier(), i)
			continue
		}
		if originalQualifier != aliasSuggestion {
			imprt.Alias = aliasSuggestion
		}
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
