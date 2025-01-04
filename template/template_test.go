package template

import (
	"go/types"
	"testing"
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
		f := TemplateMockFuncs["importStatement"].(func(*Package) string)
		pkg := NewPackage(types.NewPackage("xyz", "xyz"))
		if f(pkg) != `"xyz"` {
			t.Errorf("ImportStatement(...): want: `\"xyz\"`; got: `%s`", f(pkg))
		}

		pkg.Alias = "x"
		if f(pkg) != `x "xyz"` {
			t.Errorf("ImportStatement(...): want: `x \"xyz\"`; got: `%s`", f(pkg))
		}
	})

	t.Run("SyncPkgQualifier", func(t *testing.T) {
		f := TemplateMockFuncs["syncPkgQualifier"].(func([]*Package) string)
		if f(nil) != "sync" {
			t.Errorf("SyncPkgQualifier(...): want: `sync`; got: `%s`", f(nil))
		}
		imports := []*Package{
			NewPackage(types.NewPackage("sync", "sync")),
			NewPackage(types.NewPackage("github.com/some/module", "module")),
		}
		if f(imports) != "sync" {
			t.Errorf("SyncPkgQualifier(...): want: `sync`; got: `%s`", f(imports))
		}

		syncPkg := NewPackage(types.NewPackage("sync", "sync"))
		syncPkg.Alias = "stdsync"
		otherSyncPkg := NewPackage(types.NewPackage("github.com/someother/sync", "sync"))
		imports = []*Package{otherSyncPkg, syncPkg}
		if f(imports) != "stdsync" {
			t.Errorf("SyncPkgQualifier(...): want: `stdsync`; got: `%s`", f(imports))
		}
	})
}
