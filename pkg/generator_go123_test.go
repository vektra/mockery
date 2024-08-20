//go:build go1.23

package pkg

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"regexp"
)

func (s *GeneratorSuite) TestReplaceTypePackagePrologueGo123() {
	if !isTypeAliasEnabled() {
		// "go 1.22" in go.mod makes gotypesalias=0 even when compiling with Go 1.23.
		// Remove this when upgrading to Go 1.23 in go.mod.
		return
	}

	expected := `package mocks

import baz "github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz"
import mock "github.com/stretchr/testify/mock"

`
	generator := NewGenerator(
		s.ctx,
		GeneratorConfig{InPackage: false},
		s.getInterfaceFromFile("example_project/baz/foo.go", "Foo"),
		pkg,
	)

	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestReplaceTypePackageGo123() {
	if !isTypeAliasEnabled() {
		// "go 1.22" in go.mod makes gotypesalias=0 even when compiling with Go 1.23.
		// Remove this when upgrading to Go 1.23 in go.mod.
		return
	}

	cfg := GeneratorConfig{InPackage: false}

	s.checkGenerationRegexWithConfig("example_project/baz/foo.go", "Foo", cfg, []regexpExpected{
		// func (_m *Foo) GetBaz() (*baz.Baz, error)
		{true, regexp.MustCompile(`func \([^\)]+\) GetBaz\(\) \(\*baz\.Baz`)},
		// func (_m *Foo) GetBaz() (*foo.InternalBaz, error)
		{false, regexp.MustCompile(`func \([^\)]+\) GetBaz\(\) \(\*foo\.InternalBaz`)},
	})
}

func (s *GeneratorSuite) TestReplaceTypePackageMultiplePrologueGo123() {
	if !isTypeAliasEnabled() {
		// "go 1.22" in go.mod makes gotypesalias=0 even when compiling with Go 1.23.
		// Remove this when upgrading to Go 1.23 in go.mod.
		return
	}

	expected := `package mocks

import mock "github.com/stretchr/testify/mock"
import replace_type "github.com/vektra/mockery/v2/pkg/fixtures/example_project/replace_type"
import rt1 "github.com/vektra/mockery/v2/pkg/fixtures/example_project/replace_type/rti/rt1"
import rt2 "github.com/vektra/mockery/v2/pkg/fixtures/example_project/replace_type/rti/rt2"

`
	generator := NewGenerator(
		s.ctx,
		GeneratorConfig{InPackage: false},
		s.getInterfaceFromFile("example_project/replace_type/rt.go", "RType"),
		pkg,
	)

	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestReplaceTypePackageMultipleGo123() {
	if !isTypeAliasEnabled() {
		// "go 1.22" in go.mod makes gotypesalias=0 even when compiling with Go 1.23.
		// Remove this when upgrading to Go 1.23 in go.mod.
		return
	}

	cfg := GeneratorConfig{InPackage: false}

	s.checkGenerationRegexWithConfig("example_project/replace_type/rt.go", "RType", cfg, []regexpExpected{
		// func (_m *RType) Replace1(f rt1.RType1)
		{true, regexp.MustCompile(`func \([^\)]+\) Replace1\(f rt1\.RType1`)},
		// func (_m *RType) Replace2(f rt2.RType2)
		{true, regexp.MustCompile(`func \([^\)]+\) Replace2\(f rt2\.RType2`)},
	})
}

// isTypeAliasEnabled reports whether [NewAlias] should create [types.Alias] types.
//
// This function is expensive! Call it sparingly.
// source: /go/1.23.0/libexec/src/cmd/vendor/golang.org/x/tools/internal/aliases/aliases_go122.go
func isTypeAliasEnabled() bool {
	// The only reliable way to compute the answer is to invoke go/types.
	// We don't parse the GODEBUG environment variable, because
	// (a) it's tricky to do so in a manner that is consistent
	//     with the godebug package; in particular, a simple
	//     substring check is not good enough. The value is a
	//     rightmost-wins list of options. But more importantly:
	// (b) it is impossible to detect changes to the effective
	//     setting caused by os.Setenv("GODEBUG"), as happens in
	//     many tests. Therefore any attempt to cache the result
	//     is just incorrect.
	fset := token.NewFileSet()
	f, _ := parser.ParseFile(fset, "a.go", "package p; type A = int", 0)
	pkg, _ := new(types.Config).Check("p", fset, []*ast.File{f}, nil)
	_, enabled := pkg.Scope().Lookup("A").Type().(*types.Alias)
	return enabled
}
