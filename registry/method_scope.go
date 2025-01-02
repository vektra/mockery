package registry

import (
	"context"
	"fmt"
	"go/types"

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
	// visibleNames contains a collection of all names visible to this lexical
	// scope. This includes import qualifiers, type names etc. This is used to prevent naming
	// collisions.
	visibleNames map[string]any
	imports      map[string]*Package
}

func NewMethodScope(r *Registry) *MethodScope {
	m := &MethodScope{
		registry:     r,
		vars:         []*Var{},
		conflicted:   map[string]bool{},
		visibleNames: map[string]any{},
		imports:      map[string]*Package{},
	}
	for key := range r.importQualifiers {
		m.AddName(key)
	}
	return m
}

func (m *MethodScope) ResolveVariableNameCollisions(ctx context.Context) {
	log := zerolog.Ctx(ctx)
	for _, v := range m.vars {
		varLog := log.With().Str("variable-name", v.Name).Logger()
		newName := m.AllocateName(v.Name)
		if newName != v.Name {
			varLog.Debug().Str("new-name", newName).Msg("variable was found to conflict with previously allocated name. Giving new name.")
		}
		v.Name = newName
		m.AddName(v.Name)
	}
}

// AllocateName creates a new variable name in the lexical scope of the method.
// It ensures the returned name does not conflict with any other name visible
// to the scope. It registers the returned name in the lexical scope such that
// its exact value can never be allocated again.
func (m *MethodScope) AllocateName(prefix string) string {
	var suggestion string
	for i := 0; ; i++ {
		if i == 0 {
			suggestion = prefix
		} else {
			suggestion = fmt.Sprintf("%s%d", prefix, i)
		}

		if m.NameExists(suggestion) {
			continue
		}
		break
	}
	return suggestion
}

// AddVar allocates a variable instance and adds it to the method scope.
//
// Variables names are generated if required and are ensured to be
// without conflict with other variables and imported packages. It also
// adds the relevant imports to the registry for each added variable.
func (m *MethodScope) AddVar(vr *types.Var, prefix string) *Var {
	imports := m.populateImports(context.Background(), vr.Type())
	v := Var{
		vr:         vr,
		imports:    imports,
		moqPkgPath: m.moqPkgPath,
	}
	// The variable type is also a visible name, so add that.
	m.AddName(v.TypeString())

	v.Name = m.AllocateName(varName(vr, prefix))

	m.vars = append(m.vars, &v)
	return &v
}

// AddName records name as visible in the current scope. This may be useful
// in cases where a template statically adds its own name that needs to be registered
// with the scope to prevent future naming collisions.
func (m *MethodScope) AddName(name string) {
	m.visibleNames[name] = nil
}

// NameExists returns whether or not the name is currently visible in the scope.
func (m *MethodScope) NameExists(name string) bool {
	_, exists := m.visibleNames[name]
	return exists
}

func (m *MethodScope) addImport(ctx context.Context, pkg *types.Package, imports map[string]*Package) {
	imprt := m.registry.AddImport(ctx, pkg)
	imports[pkg.Path()] = imprt
	m.imports[pkg.Path()] = imprt
	m.AddName(imprt.Qualifier())
}

func (m *MethodScope) populateImportNamedType(
	ctx context.Context,
	t interface {
		Obj() *types.TypeName
		TypeArgs() *types.TypeList
	},
	imports map[string]*Package,
) {
	if pkg := t.Obj().Pkg(); pkg != nil {
		m.addImport(ctx, pkg, imports)
	}
	// The imports of a Type with a TypeList must be added to the imports list
	// For example: Foo[otherpackage.Bar] , must have otherpackage imported
	if targs := t.TypeArgs(); targs != nil {
		for i := 0; i < targs.Len(); i++ {
			m.populateImportsHelper(ctx, targs.At(i), imports)
		}
	}
}

func (m *MethodScope) populateImports(ctx context.Context, t types.Type) map[string]*Package {
	imports := map[string]*Package{}
	m.populateImportsHelper(ctx, t, imports)
	return imports
}

// populateImportsHelper extracts all the package imports for a given type
// recursively. The imported packages by a single type can be more than
// one (ex: map[a.Type]b.Type).
//
// Returned are the imports that were added for the given type.
func (m *MethodScope) populateImportsHelper(ctx context.Context, t types.Type, imports map[string]*Package) {
	log := zerolog.Ctx(ctx).With().
		Str("type-str", t.String()).Logger()
	switch t := t.(type) {
	case *types.Named:
		m.populateImportNamedType(ctx, t, imports)
	case *types.Alias:
		m.populateImportNamedType(ctx, t, imports)
	case *types.Array:
		m.populateImportsHelper(ctx, t.Elem(), imports)

	case *types.Slice:
		m.populateImportsHelper(ctx, t.Elem(), imports)

	case *types.Signature:
		for i := 0; i < t.Params().Len(); i++ {
			m.populateImportsHelper(ctx, t.Params().At(i).Type(), imports)
		}
		for i := 0; i < t.Results().Len(); i++ {
			m.populateImportsHelper(ctx, t.Results().At(i).Type(), imports)
		}

	case *types.Map:
		m.populateImportsHelper(ctx, t.Key(), imports)
		m.populateImportsHelper(ctx, t.Elem(), imports)

	case *types.Chan:
		m.populateImportsHelper(ctx, t.Elem(), imports)

	case *types.Pointer:
		m.populateImportsHelper(ctx, t.Elem(), imports)

	case *types.Struct: // anonymous struct
		for i := 0; i < t.NumFields(); i++ {
			m.populateImportsHelper(ctx, t.Field(i).Type(), imports)
		}

	case *types.Union:
		log.Debug().Int("len", t.Len()).Msg("found union")
		for i := 0; i < t.Len(); i++ {
			term := t.Term(i)
			m.populateImportsHelper(ctx, term.Type(), imports)
		}
	case *types.Interface: // anonymous interface
		log.Debug().
			Int("num-methods", t.NumMethods()).
			Int("num-explicit-methods", t.NumExplicitMethods()).
			Int("num-embeddeds", t.NumEmbeddeds()).
			Msg("found interface")
		for i := 0; i < t.NumExplicitMethods(); i++ {
			log.Debug().Msg("populating import from explicit method")
			m.populateImportsHelper(ctx, t.ExplicitMethod(i).Type(), imports)
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			log.Debug().Msg("populating import form embedded type")
			m.populateImportsHelper(ctx, t.EmbeddedType(i), imports)
		}
	case *types.Basic:
		if t.Kind() == types.UnsafePointer {
			m.addImport(ctx, types.Unsafe, imports)
		}
	default:
		log.Debug().Str("real-type", fmt.Sprintf("%T", t)).Msg("unable to determine type of object")
	}
}
