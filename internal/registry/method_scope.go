package registry

import (
	"go/types"
	"strconv"
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
func (m *MethodScope) AddVar(vr *types.Var, suffix string) *Var {
	name := vr.Name()
	if name == "" || name == "_" {
		name = generateVarName(vr.Type())
	}

	name += suffix

	switch name {
	case "mock", "callInfo", "break", "default", "func", "interface", "select", "case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch", "const", "fallthrough", "if", "range", "type", "continue", "for",
		"import", "return", "var":
		name += "MoqParam"
	}

	if _, ok := m.searchVar(name); ok || m.conflicted[name] {
		return m.addDisambiguatedVar(vr, name)
	}

	return m.addVar(vr, name)
}

func (m *MethodScope) addDisambiguatedVar(vr *types.Var, suggested string) *Var {
	n := 1
	for {
		// Keep incrementing the suffix until we find a name which is unused.
		if _, ok := m.searchVar(suggested + strconv.Itoa(n)); !ok {
			break
		}
		n++
	}

	name := suggested + strconv.Itoa(n)
	if n == 1 {
		conflict, _ := m.searchVar(suggested)
		conflict.Name += "1"
		name = suggested + "2"
		m.conflicted[suggested] = true
	}

	return m.addVar(vr, name)
}

func (m *MethodScope) addVar(vr *types.Var, name string) *Var {
	imports := make(map[string]*Package)
	m.populateImports(vr.Type(), imports)

	v := Var{
		vr:         vr,
		imports:    imports,
		moqPkgPath: m.moqPkgPath,
		Name:       name,
	}
	m.vars = append(m.vars, &v)
	m.resolveImportVarConflicts(&v)
	return &v
}

func (m MethodScope) searchVar(name string) (*Var, bool) {
	for _, v := range m.vars {
		if v.Name == name {
			return v, true
		}
	}

	return nil, false
}

// populateImports extracts all the package imports for a given type
// recursively. The imported packages by a single type can be more than
// one (ex: map[a.Type]b.Type).
func (m MethodScope) populateImports(t types.Type, imports map[string]*Package) {
	switch t := t.(type) {
	case *types.Named:
		if pkg := t.Obj().Pkg(); pkg != nil {
			imports[stripVendorPath(pkg.Path())] = m.registry.AddImport(pkg)
		}

	case *types.Array:
		m.populateImports(t.Elem(), imports)

	case *types.Slice:
		m.populateImports(t.Elem(), imports)

	case *types.Signature:
		for i := 0; i < t.Params().Len(); i++ {
			m.populateImports(t.Params().At(i).Type(), imports)
		}
		for i := 0; i < t.Results().Len(); i++ {
			m.populateImports(t.Results().At(i).Type(), imports)
		}

	case *types.Map:
		m.populateImports(t.Key(), imports)
		m.populateImports(t.Elem(), imports)

	case *types.Chan:
		m.populateImports(t.Elem(), imports)

	case *types.Pointer:
		m.populateImports(t.Elem(), imports)

	case *types.Struct: // anonymous struct
		for i := 0; i < t.NumFields(); i++ {
			m.populateImports(t.Field(i).Type(), imports)
		}

	case *types.Interface: // anonymous interface
		for i := 0; i < t.NumExplicitMethods(); i++ {
			m.populateImports(t.ExplicitMethod(i).Type(), imports)
		}
		for i := 0; i < t.NumEmbeddeds(); i++ {
			m.populateImports(t.EmbeddedType(i), imports)
		}
	}
}

func (m MethodScope) resolveImportVarConflicts(v *Var) {
	// Ensure that the newly added var does not conflict with a package import
	// which was added earlier.
	if _, ok := m.registry.searchImport(v.Name); ok {
		v.Name += "MoqParam"
	}
	// Ensure that all the newly added imports do not conflict with any of the
	// existing vars.
	for _, imprt := range v.imports {
		if v, ok := m.searchVar(imprt.Qualifier()); ok {
			v.Name += "MoqParam"
		}
	}
}
