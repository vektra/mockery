package internal

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

type fakeImporter map[string]*types.Package

func (f fakeImporter) Import(path string) (*types.Package, error) {
	if p, ok := f[path]; ok {
		return p, nil
	}
	return importer.Default().Import(path)
}

func check(t *testing.T, imp fakeImporter, path, src string) *packages.Package {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path+".go", src, 0)
	require.NoError(t, err)
	pkg, err := (&types.Config{Importer: imp}).Check(path, fset, []*ast.File{f}, nil)
	require.NoError(t, err)
	imp[path] = pkg
	return &packages.Package{PkgPath: path, Types: pkg}
}

func TestAliasIndexResolve(t *testing.T) {
	imp := fakeImporter{}
	internalPkg := check(t, imp, "example.com/lib/internal", `package internal
type Foo struct{}
type Bar struct{}
type Orphan struct{}
`)
	// Two packages alias Foo; the first alphabetically (a) wins. Bar is renamed.
	short := check(t, imp, "example.com/lib/a", `package a
import "example.com/lib/internal"
type Foo = internal.Foo
type Renamed = internal.Bar
`)
	long := check(t, imp, "example.com/lib/longer", `package longer
import "example.com/lib/internal"
type Foo = internal.Foo
`)
	check(t, imp, "example.com/lib/box", `package box
type Box[T any] struct{}
`)
	user := check(t, imp, "example.com/lib/user", `package user
import (
	"example.com/lib/box"
	"example.com/lib/internal"
)
type T struct{}
func (T) Nested(f func(x *internal.Foo), b ...internal.Bar) map[string][]internal.Foo { return nil }
func (T) Orphan() internal.Orphan { return internal.Orphan{} }
func (T) Generic(b box.Box[internal.Foo]) box.Box[internal.Bar] { return box.Box[internal.Bar]{} }
func (T) Struct(s struct {
	F internal.Foo
	N int
}) {
}
func (T) Iface(i interface{ M(internal.Foo) }) {}
func (T) OrphanInStruct(s struct{ O internal.Orphan }) {}
type G[X ~int | internal.Foo] interface{}
`)

	// An alias that lives inside an internal package is no use to outside code.
	hidden := check(t, imp, "example.com/lib/internal/hidden", `package hidden
import "example.com/lib/internal"
type Hidden = internal.Orphan
`)

	idx := AliasIndex{}
	for _, p := range []*packages.Package{internalPkg, long, hidden, short} {
		idx.parse(p)
	}

	methods := map[string]*types.Signature{}
	named := user.Types.Scope().Lookup("T").Type().(*types.Named)
	for i := 0; i < named.NumMethods(); i++ {
		methods[named.Method(i).Name()] = named.Method(i).Type().(*types.Signature)
	}
	qual := func(p *types.Package) string { return p.Name() }

	// From outside example.com/lib the internal package is off limits, so
	// types are swapped, including ones nested in func params, variadics,
	// maps, slices and generic arguments.
	const outside = "example.com/other/mocks"
	got, err := idx.resolveType(methods["Nested"], outside)
	require.NoError(t, err)
	require.Equal(t, "func(f func(x *a.Foo), b ...a.Renamed) map[string][]a.Foo", types.TypeString(got, qual))

	got, err = idx.resolveType(methods["Generic"], outside)
	require.NoError(t, err)
	require.Equal(t, "func(b box.Box[a.Foo]) box.Box[a.Renamed]", types.TypeString(got, qual))

	// Struct and interface literals are walked too.
	got, err = idx.resolveType(methods["Struct"], outside)
	require.NoError(t, err)
	require.Equal(t, "func(s struct{F a.Foo; N int})", types.TypeString(got, qual))

	got, err = idx.resolveType(methods["Iface"], outside)
	require.NoError(t, err)
	require.Equal(t, "func(i interface{M(a.Foo)})", types.TypeString(got, qual))

	_, err = idx.resolveType(methods["OrphanInStruct"], outside)
	require.ErrorContains(t, err, "internal.Orphan")

	// So are union constraints on a generic interface's type parameters.
	constraint := user.Types.Scope().Lookup("G").Type().(*types.Named).TypeParams().At(0).Constraint()
	got, err = idx.resolveType(constraint, outside)
	require.NoError(t, err)
	require.Equal(t, "~int | a.Foo", types.TypeString(got, qual))
	require.True(t, got.(*types.Interface).IsImplicit(), "implicit interface stays implicit")

	// Orphan has no alias and can't be imported from outside: that's an error.
	_, err = idx.resolveType(methods["Orphan"], outside)
	require.ErrorContains(t, err, "internal.Orphan")

	// From inside example.com/lib, internal is importable: nothing is swapped, nothing errors.
	const inside = "example.com/lib/mocks"
	for name, sig := range methods {
		got, err := idx.resolveType(sig, inside)
		require.NoError(t, err, name)
		require.Same(t, sig, got, name)
	}

	// A nil index means the option is off: no swap and no error.
	var off AliasIndex
	got, err = off.resolveType(methods["Orphan"], outside)
	require.NoError(t, err)
	require.Same(t, methods["Orphan"], got)

	// Nothing to replace -> same value back.
	plain := types.Typ[types.String]
	got, err = idx.resolveType(plain, outside)
	require.NoError(t, err)
	require.Same(t, plain, got)
}

func TestImportable(t *testing.T) {
	for _, c := range []struct {
		from, path string
		want       bool
	}{
		{"github.com/me/mocks", "go.temporal.io/sdk/client", true},
		{"github.com/me/mocks", "go.temporal.io/sdk/internal", false},
		{"github.com/me/mocks", "go.temporal.io/sdk/internal/common", false},
		{"go.temporal.io/sdk/mocks", "go.temporal.io/sdk/internal/common", true},
		{"go.temporal.io/sdk", "go.temporal.io/sdk/internal", true},
		{"go.temporal.io/sdkx/mocks", "go.temporal.io/sdk/internal", false}, // prefix must end at a path element
		{"a/b/internal/mocks", "a/b/internal/c/internal/d", false},          // the last internal element decides
		{"a/b/internal/c/mocks", "a/b/internal/c/internal/d", true},
		{"example.com/me", "internal/abi", false}, // the standard library's
	} {
		require.Equal(t, c.want, importable(c.from, c.path), "%s -> %s", c.from, c.path)
	}
}

func TestIsInternalPath(t *testing.T) {
	for path, want := range map[string]bool{
		"internal":                           true,
		"internal/foo":                       true,
		"go.temporal.io/sdk/internal":        true,
		"go.temporal.io/sdk/internal/common": true,
		"/internal/":                         true, // not a valid import path, but must not misbehave
		"go.temporal.io/sdk/client":          false,
		"example.com/internalstuff":          false,
		"example.com/myinternal":             false,
		"example.com/internal2/foo":          false,
		"":                                   false,
	} {
		require.Equal(t, want, isInternalPath(path), path)
	}
}
