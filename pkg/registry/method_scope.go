package registry

import (
	"context"
	"go/types"
	"strconv"

	"github.com/rs/zerolog"
)

// MethodScope is the sub-registry for allocating variables present in
// the method scope.
//
// It should be created using a registry instance.
type MethodScope struct {
	registry   *Registry
	moqPkgPath string

	vars       []*Var
	conflicted map[string]bool
}

// AddVar allocates a variable instance and adds it to the method scope.
//
// Variables names are generated if required and are ensured to be
// without conflict with other variables and imported packages. It also
// adds the relevant imports to the registry for each added variable.
func (m *MethodScope) AddVar(ctx context.Context, vr *types.Var, suffix string) *Var {
	imports := make(map[string]*Package)
	m.populateImports(ctx, vr.Type(), imports)
	m.resolveImportVarConflicts(imports)

	name := varName(vr, suffix)
	// Ensure that the var name does not conflict with a package import.
	if _, ok := m.registry.searchImport(name); ok {
		name += "MoqParam"
	}
	if _, ok := m.searchVar(name); ok || m.conflicted[name] {
		name = m.resolveVarNameConflict(name)
	}

	v := Var{
		vr:         vr,
		imports:    imports,
		moqPkgPath: m.moqPkgPath,
		Name:       name,
	}
	m.vars = append(m.vars, &v)
	return &v
}

func (m *MethodScope) resolveVarNameConflict(suggested string) string {
	for n := 1; ; n++ {
		_, ok := m.searchVar(suggested + strconv.Itoa(n))
		if ok {
			continue
		}

		if n == 1 {
			conflict, _ := m.searchVar(suggested)
			conflict.Name += "1"
			m.conflicted[suggested] = true
			n++
		}
		return suggested + strconv.Itoa(n)
	}
}

func (m MethodScope) searchVar(name string) (*Var, bool) {
	for _, v := range m.vars {
		if v.Name == name {
			return v, true
		}
	}

	return nil, false
}

func (m MethodScope) populateImportNamedType(
	ctx context.Context,
	t interface {
		Obj() *types.TypeName
		TypeArgs() *types.TypeList
	},
	imports map[string]*Package,
) {
	if pkg := t.Obj().Pkg(); pkg != nil {
		imports[pkg.Path()] = m.registry.AddImport(ctx, pkg)
	}
	// The imports of a Type with a TypeList must be added to the imports list
	// For example: Foo[otherpackage.Bar] , must have otherpackage imported
	if targs := t.TypeArgs(); targs != nil {
		for i := 0; i < targs.Len(); i++ {
			m.populateImports(ctx, targs.At(i), imports)
		}
	}
}

// populateImports extracts all the package imports for a given type
// recursively. The imported packages by a single type can be more than
// one (ex: map[a.Type]b.Type).
func (m MethodScope) populateImports(ctx context.Context, t types.Type, imports map[string]*Package) {
	log := zerolog.Ctx(ctx).With().
		Str("type-str", t.String()).Logger()
	switch t := t.(type) {
	case *types.Named:
		m.populateImportNamedType(ctx, t, imports)
	case *types.Alias:
		m.populateImportNamedType(ctx, t, imports)
	case *types.Array:
		m.populateImports(ctx, t.Elem(), imports)

	case *types.Slice:
		m.populateImports(ctx, t.Elem(), imports)

	case *types.Signature:
		for i := 0; i < t.Params().Len(); i++ {
			m.populateImports(ctx, t.Params().At(i).Type(), imports)
		}
		for i := 0; i < t.Results().Len(); i++ {
			m.populateImports(ctx, t.Results().At(i).Type(), imports)
		}

	case *types.Map:
		m.populateImports(ctx, t.Key(), imports)
		m.populateImports(ctx, t.Elem(), imports)

	case *types.Chan:
		m.populateImports(ctx, t.Elem(), imports)

	case *types.Pointer:
		m.populateImports(ctx, t.Elem(), imports)

	case *types.Struct: // anonymous struct
		for i := 0; i < t.NumFields(); i++ {
			m.populateImports(ctx, t.Field(i).Type(), imports)
		}

	case *types.Interface: // anonymous interface
		for i := 0; i < t.NumExplicitMethods(); i++ {
			m.populateImports(ctx, t.ExplicitMethod(i).Type(), imports)
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			m.populateImports(ctx, t.EmbeddedType(i), imports)
		}
	default:
		log.Debug().Msg("unable to determine type of object")
	}
}

// resolveImportVarConflicts ensures that all the newly added imports do not
// conflict with any of the existing vars.
func (m MethodScope) resolveImportVarConflicts(imports map[string]*Package) {
	// Ensure that all the newly added imports do not conflict with any of the
	// existing vars.
	for _, imprt := range imports {
		if v, ok := m.searchVar(imprt.Qualifier()); ok {
			v.Name += "MoqParam"
		}
	}
}
