package template

import (
	"go/types"
	"testing"

	"github.com/vektra/mockery/v3/pkg/registry"
)

func TestTemplateMockFuncs(t *testing.T) {
	t.Run("Exported", func(t *testing.T) {
		f := TemplateMockFuncs["exported"].(func(string) string)
		if f("") != "" {
			t.Errorf("Exported(...) want: ``; got: `%s`", f(""))
		}
		if f("var") != "Var" {
			t.Errorf("Exported(...) want: `Var`; got: `%s`", f("var"))
		}
	})

	t.Run("ImportStatement", func(t *testing.T) {
		f := TemplateMockFuncs["ImportStatement"].(func(*registry.Package) string)
		pkg := registry.NewPackage(types.NewPackage("xyz", "xyz"))
		if f(pkg) != `"xyz"` {
			t.Errorf("ImportStatement(...): want: `\"xyz\"`; got: `%s`", f(pkg))
		}

		pkg.Alias = "x"
		if f(pkg) != `x "xyz"` {
			t.Errorf("ImportStatement(...): want: `x \"xyz\"`; got: `%s`", f(pkg))
		}
	})

	t.Run("SyncPkgQualifier", func(t *testing.T) {
		f := TemplateMockFuncs["SyncPkgQualifier"].(func([]*registry.Package) string)
		if f(nil) != "sync" {
			t.Errorf("SyncPkgQualifier(...): want: `sync`; got: `%s`", f(nil))
		}
		imports := []*registry.Package{
			registry.NewPackage(types.NewPackage("sync", "sync")),
			registry.NewPackage(types.NewPackage("github.com/some/module", "module")),
		}
		if f(imports) != "sync" {
			t.Errorf("SyncPkgQualifier(...): want: `sync`; got: `%s`", f(imports))
		}

		syncPkg := registry.NewPackage(types.NewPackage("sync", "sync"))
		syncPkg.Alias = "stdsync"
		otherSyncPkg := registry.NewPackage(types.NewPackage("github.com/someother/sync", "sync"))
		imports = []*registry.Package{otherSyncPkg, syncPkg}
		if f(imports) != "stdsync" {
			t.Errorf("SyncPkgQualifier(...): want: `stdsync`; got: `%s`", f(imports))
		}
	})
}
