package internal

import (
	"fmt"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// AliasIndex maps "pkg/path.TypeName" of a type declared in an internal package
// to every declared alias of that type, sorted for determinism.
//
// Libraries often declare their API in an internal package and re-export it:
//
//	package client
//	type StartWorkflowOptions = internal.StartWorkflowOptions
//
// Method signatures are written in terms of the original (internal) type, which
// code outside the module is not allowed to import. When
// `resolve-internal-aliases` is enabled, we render the public alias instead.
type AliasIndex map[string][]*types.TypeName

// internalRoot returns the parent of the last `internal` element in path, which is
// the root of the tree that may import it. ok is false if path has no such element.
func internalRoot(path string) (root string, ok bool) {
	// Wrapping the path in slashes lets one search cover an `internal` element
	// at the start, middle, end, or the whole path.
	p := "/" + path + "/"
	i := strings.LastIndex(p, "/internal/")
	if i < 0 {
		return "", false
	}
	if i == 0 {
		return "", true // a top-level `internal/...` belongs to the standard library
	}
	return p[1:i], true
}

// isInternalPath reports whether an import path contains an `internal` element.
func isInternalPath(path string) bool {
	_, ok := internalRoot(path)
	return ok
}

// importable reports whether code in the package `from` may import `path`.
func importable(from, path string) bool {
	root, ok := internalRoot(path)
	if !ok {
		return true
	}
	return root != "" && (from == root || strings.HasPrefix(from, root+"/"))
}

func typeKey(obj *types.TypeName) string {
	return obj.Pkg().Path() + "." + obj.Name()
}

// parse records every exported alias in pkg whose target is a named type in an
// internal package. Every alias is kept, sorted by name and exported so the results are
// deterministic. Whether an alias can actually be used is decided later, per
// output package, in resolveType, which picks the first importable one.
func (a AliasIndex) parse(pkg *packages.Package) {
	// Skip packages with no type info.
	if a == nil || pkg.Types == nil {
		return
	}
	// Go over every name declared at the top level of the package.
	scope := pkg.Types.Scope()
	for _, name := range scope.Names() {
		// Must be exported and an alias otherwise skip.
		tn, ok := scope.Lookup(name).(*types.TypeName)
		if !ok || !tn.Exported() || !tn.IsAlias() {
			continue
		}
		// Follow the alias to see if it's internal, if not skip.
		target, ok := types.Unalias(tn.Type()).(*types.Named)
		if !ok || target.Obj().Pkg() == nil || !isInternalPath(target.Obj().Pkg().Path()) {
			continue
		}
		key := typeKey(target.Obj())
		candidates := append(a[key], tn)
		sort.Slice(candidates, func(i, j int) bool {
			return replaceAlias(candidates[i], candidates[j])
		})
		a[key] = candidates
	}
}

// replaceAlias reports whether candidate should replace current as the alias for a type.
// An exported alias wins over an internal one to favor ones you can use.
// Then sorts alphabetically to be deterministic.
func replaceAlias(candidate, current *types.TypeName) bool {
	otherInternal := isInternalPath(candidate.Pkg().Path())
	currentInternal := isInternalPath(current.Pkg().Path())
	if otherInternal != currentInternal {
		// If the candidate is not internal and the current isn't, it should replace the current alias.
		return !otherInternal
	}
	return candidate.Type().String() < current.Type().String()
}

// resolveVar swaps the types of v that the package `from` can't import for their public
// aliases, and errors if one has no alias. A nil index (option off) returns v unchanged.
func (a AliasIndex) resolveVar(v *types.Var, from string) (*types.Var, error) {
	t, err := a.resolveType(v.Type(), from)
	if err != nil {
		return nil, err
	}
	if t != v.Type() {
		return types.NewVar(v.Pos(), v.Pkg(), v.Name(), t), nil
	}
	return v, nil
}

// resolveType swaps types that the package `from` can't import for their public aliases,
// walking into pointers, slices, maps, channels, generic arguments, func signatures,
// struct and interface literals, and unions.
// A type that can't be imported and has no alias is an error. If there is nothing to
// swap, it returns t unchanged. A nil index (option off) never swaps or errors.
func (a AliasIndex) resolveType(t types.Type, from string) (types.Type, error) {
	if a == nil {
		// Persists existing behavior if alias' aren't resolved.
		// But means the internal types are compile errors and not reported by mockery.
		return t, nil
	}
	switch t := t.(type) {
	case *types.Alias:
		// An alias the output package can import is rendered as written.
		if pkg := t.Obj().Pkg(); pkg == nil || importable(from, pkg.Path()) {
			return t, nil
		}
		// Otherwise look through it to the type it stands for.
		return a.resolveType(types.Unalias(t), from)
	case *types.Named:
		pkg := t.Obj().Pkg()
		args := t.TypeArgs()
		if pkg != nil && !importable(from, pkg.Path()) {
			// Use the first alias whose package the output package may import too.
			if args.Len() == 0 {
				for _, alias := range a[typeKey(t.Obj())] {
					if importable(from, alias.Pkg().Path()) {
						return alias.Type(), nil
					}
				}
			}
			return nil, fmt.Errorf("%s.%s is declared in %q, which %q can't import, and no public alias for it was found",
				pkg.Name(), t.Obj().Name(), pkg.Path(), from)
		}
		// An importable generic type, e.g. Box[internal.Foo]: resolve its arguments.
		newArgs := make([]types.Type, args.Len())
		changed := false
		for i := range newArgs {
			na, err := a.resolveType(args.At(i), from)
			if err != nil {
				return nil, err
			}
			newArgs[i], changed = na, changed || na != args.At(i)
		}
		if changed {
			return types.Instantiate(nil, t.Origin(), newArgs, false)
		}
	case *types.Pointer:
		e, err := a.resolveType(t.Elem(), from)
		if err != nil {
			return nil, err
		}
		if e != t.Elem() {
			return types.NewPointer(e), nil
		}
	case *types.Slice:
		e, err := a.resolveType(t.Elem(), from)
		if err != nil {
			return nil, err
		}
		if e != t.Elem() {
			return types.NewSlice(e), nil
		}
	case *types.Array:
		e, err := a.resolveType(t.Elem(), from)
		if err != nil {
			return nil, err
		}
		if e != t.Elem() {
			return types.NewArray(e, t.Len()), nil
		}
	case *types.Chan:
		e, err := a.resolveType(t.Elem(), from)
		if err != nil {
			return nil, err
		}
		if e != t.Elem() {
			return types.NewChan(t.Dir(), e), nil
		}
	case *types.Map:
		k, err := a.resolveType(t.Key(), from)
		if err != nil {
			return nil, err
		}
		e, err := a.resolveType(t.Elem(), from)
		if err != nil {
			return nil, err
		}
		if k != t.Key() || e != t.Elem() {
			return types.NewMap(k, e), nil
		}
	case *types.Struct:
		fields := make([]*types.Var, t.NumFields())
		tags := make([]string, t.NumFields())
		changed := false
		for i := range fields {
			f := t.Field(i)
			nt, err := a.resolveType(f.Type(), from)
			if err != nil {
				return nil, err
			}
			if nt != f.Type() {
				f = types.NewField(f.Pos(), f.Pkg(), f.Name(), nt, f.Embedded())
				changed = true
			}
			fields[i], tags[i] = f, t.Tag(i)
		}
		if changed {
			return types.NewStruct(fields, tags), nil
		}
	case *types.Interface:
		methods := make([]*types.Func, t.NumExplicitMethods())
		embeddeds := make([]types.Type, t.NumEmbeddeds())
		changed := false
		for i := range methods {
			m := t.ExplicitMethod(i)
			sig, err := a.resolveType(m.Type(), from)
			if err != nil {
				return nil, err
			}
			if sig != m.Type() {
				m = types.NewFunc(m.Pos(), m.Pkg(), m.Name(), sig.(*types.Signature))
				changed = true
			}
			methods[i] = m
		}
		for i := range embeddeds {
			e, err := a.resolveType(t.EmbeddedType(i), from)
			if err != nil {
				return nil, err
			}
			embeddeds[i], changed = e, changed || e != t.EmbeddedType(i)
		}
		if changed {
			it := types.NewInterfaceType(methods, embeddeds)
			if t.IsImplicit() {
				it.MarkImplicit()
			}
			return it.Complete(), nil
		}
	case *types.Union:
		terms := make([]*types.Term, t.Len())
		changed := false
		for i := range terms {
			term := t.Term(i)
			nt, err := a.resolveType(term.Type(), from)
			if err != nil {
				return nil, err
			}
			if nt != term.Type() {
				term = types.NewTerm(term.Tilde(), nt)
				changed = true
			}
			terms[i] = term
		}
		if changed {
			return types.NewUnion(terms), nil
		}
	case *types.Signature:
		if t.TypeParams().Len() > 0 {
			return t, nil // generic signatures are left alone
		}
		params, pc, err := a.resolveTuple(t.Params(), from)
		if err != nil {
			return nil, err
		}
		results, rc, err := a.resolveTuple(t.Results(), from)
		if err != nil {
			return nil, err
		}
		if pc || rc {
			return types.NewSignatureType(t.Recv(), nil, nil, params, results, t.Variadic()), nil
		}
	}
	return t, nil
}

func (a AliasIndex) resolveTuple(tup *types.Tuple, from string) (*types.Tuple, bool, error) {
	if tup == nil {
		return nil, false, nil
	}
	vars := make([]*types.Var, tup.Len())
	changed := false
	for i := range vars {
		v := tup.At(i)
		nt, err := a.resolveType(v.Type(), from)
		if err != nil {
			return nil, false, err
		}
		if nt != v.Type() {
			v = types.NewVar(v.Pos(), v.Pkg(), v.Name(), nt)
			changed = true
		}
		vars[i] = v
	}
	if !changed {
		return tup, false, nil
	}
	return types.NewTuple(vars...), true, nil
}
