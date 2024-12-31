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
	// scope. This includes import qualifiers. This is used to prevent naming
	// collisions.
	visibleNames map[string]any
	imports      map[string]*Package
}

func NewMethodScope(r *Registry) *MethodScope {
	visibleNames := map[string]any{}
	for key := range r.importQualifiers {
		visibleNames[key] = nil
	}
	return &MethodScope{
		registry:     r,
		vars:         []*Var{},
		conflicted:   map[string]bool{},
		visibleNames: visibleNames,
		imports:      map[string]*Package{},
	}
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
		m.visibleNames[v.Name] = nil
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

		if _, suggestionExists := m.visibleNames[suggestion]; suggestionExists {
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
func (m *MethodScope) AddVar(ctx context.Context, vr *types.Var, prefix string) *Var {
	log := zerolog.Ctx(ctx).
		With().
		Str("prefix", prefix).
		Str("variable-name", vr.Name()).
		Logger()
	imports := m.PopulateImports(ctx, vr.Type())

	log.Debug().Msg("adding var")
	for key := range m.visibleNames {
		log.Debug().Str("visible-name", key).Msg("visible name")
	}
	name := m.AllocateName(varName(vr, prefix))
	// This suggested name is subject to change because it might come into conflict
	// with a future package import.
	log.Debug().Str("suggested-name", name).Msg("suggested name for variable in method")

	v := Var{
		vr:         vr,
		imports:    imports,
		moqPkgPath: m.moqPkgPath,
		Name:       name,
	}
	m.vars = append(m.vars, &v)
	return &v
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
		imprt := m.registry.AddImport(ctx, pkg)
		imports[pkg.Path()] = imprt
		m.imports[pkg.Path()] = imprt
		m.visibleNames[imprt.Qualifier()] = nil
	}
	// The imports of a Type with a TypeList must be added to the imports list
	// For example: Foo[otherpackage.Bar] , must have otherpackage imported
	if targs := t.TypeArgs(); targs != nil {
		for i := 0; i < targs.Len(); i++ {
			m.populateImports(ctx, targs.At(i), imports)
		}
	}
}

func (m *MethodScope) PopulateImports(ctx context.Context, t types.Type) map[string]*Package {
	imports := map[string]*Package{}
	m.populateImports(ctx, t, imports)
	return imports
}

// populateImports extracts all the package imports for a given type
// recursively. The imported packages by a single type can be more than
// one (ex: map[a.Type]b.Type).
//
// Returned are the imports that were added for the given type.
func (m *MethodScope) populateImports(ctx context.Context, t types.Type, imports map[string]*Package) {
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

	case *types.Union:
		log.Debug().Int("len", t.Len()).Msg("found union")
		for i := 0; i < t.Len(); i++ {
			term := t.Term(i)
			m.populateImports(ctx, term.Type(), imports)
		}
	case *types.Interface: // anonymous interface
		log.Debug().
			Int("num-methods", t.NumMethods()).
			Int("num-explicit-methods", t.NumExplicitMethods()).
			Int("num-embeddeds", t.NumEmbeddeds()).
			Msg("found interface")
		for i := 0; i < t.NumExplicitMethods(); i++ {
			log.Debug().Msg("populating import from explicit method")
			m.populateImports(ctx, t.ExplicitMethod(i).Type(), imports)
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			log.Debug().Msg("populating import form embedded type")
			m.populateImports(ctx, t.EmbeddedType(i), imports)
		}

	default:
		log.Debug().Str("real-type", fmt.Sprintf("%T", t)).Msg("unable to determine type of object")
	}
}
