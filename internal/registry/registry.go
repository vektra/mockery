package registry

import (
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Registry encapsulates types information for the source and mock
// destination package. For the mock package, it tracks the list of
// imports and ensures there are no conflicts in the imported package
// qualifiers.
type Registry struct {
	srcPkgName  string
	srcPkgTypes *types.Package
	moqPkgPath  string
	aliases     map[string]string
	imports     map[string]*Package
}

// New loads the source package info and returns a new instance of
// Registry.
func New(srcDir, moqPkg string) (*Registry, error) {
	srcPkg, err := pkgInfoFromPath(
		srcDir, packages.NeedName|packages.NeedSyntax|packages.NeedTypes,
	)
	if err != nil {
		return nil, fmt.Errorf("couldn't load source package: %s", err)
	}

	return &Registry{
		srcPkgName:  srcPkg.Name,
		srcPkgTypes: srcPkg.Types,
		moqPkgPath:  findPkgPath(moqPkg, srcPkg.PkgPath),
		aliases:     parseImportsAliases(srcPkg.Syntax),
		imports:     make(map[string]*Package),
	}, nil
}

// SrcPkg returns the types info for the source package.
func (r Registry) SrcPkg() *types.Package {
	return r.srcPkgTypes
}

// SrcPkgName returns the name of the source package.
func (r Registry) SrcPkgName() string {
	return r.srcPkgName
}

// LookupInterface returns the underlying interface definition of the
// given interface name.
func (r Registry) LookupInterface(name string) (*types.Interface, *types.TypeParamList, error) {
	obj := r.SrcPkg().Scope().Lookup(name)
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
	return &MethodScope{
		registry:   r,
		moqPkgPath: r.moqPkgPath,
		conflicted: map[string]bool{},
	}
}

// AddImport adds the given package to the set of imports. It generates a
// suitable alias if there are any conflicts with previously imported
// packages.
func (r *Registry) AddImport(pkg *types.Package) *Package {
	path := stripVendorPath(pkg.Path())
	if path == r.moqPkgPath {
		return nil
	}

	if imprt, ok := r.imports[path]; ok {
		return imprt
	}

	imprt := Package{pkg: pkg, Alias: r.aliases[path]}

	if conflict, ok := r.searchImport(imprt.Qualifier()); ok {
		r.resolveImportConflict(&imprt, conflict, 0)
	}

	r.imports[path] = &imprt
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

func (r Registry) searchImport(name string) (*Package, bool) {
	for _, imprt := range r.imports {
		if imprt.Qualifier() == name {
			return imprt, true
		}
	}

	return nil, false
}

// resolveImportConflict generates and assigns a unique alias for
// packages with conflicting qualifiers.
func (r Registry) resolveImportConflict(a, b *Package, lvl int) {
	if a.uniqueName(lvl) == b.uniqueName(lvl) {
		r.resolveImportConflict(a, b, lvl+1)
		return
	}

	for _, p := range []*Package{a, b} {
		name := p.uniqueName(lvl)
		// Even though the name is not conflicting with the other package we
		// got, the new name we want to pick might already be taken. So check
		// again for conflicts and resolve them as well. Since the name for
		// this package would also get set in the recursive function call, skip
		// setting the alias after it.
		if conflict, ok := r.searchImport(name); ok && conflict != p {
			r.resolveImportConflict(p, conflict, lvl+1)
			continue
		}

		p.Alias = name
	}
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

func findPkgPath(pkgInputVal string, srcPkgPath string) string {
	if pkgInputVal == "" {
		return srcPkgPath
	}
	if pkgInDir(srcPkgPath, pkgInputVal) {
		return srcPkgPath
	}
	subdirectoryPath := filepath.Join(srcPkgPath, pkgInputVal)
	if pkgInDir(subdirectoryPath, pkgInputVal) {
		return subdirectoryPath
	}
	return ""
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
