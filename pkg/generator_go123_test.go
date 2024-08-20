//go:build go1.23

package pkg

import "regexp"

func (s *GeneratorSuite) TestReplaceTypePackagePrologueGo123() {
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
	cfg := GeneratorConfig{InPackage: false}

	s.checkGenerationRegexWithConfig("example_project/baz/foo.go", "Foo", cfg, []regexpExpected{
		// func (_m *Foo) GetBaz() (*baz.Baz, error)
		{true, regexp.MustCompile(`func \([^\)]+\) GetBaz\(\) \(\*baz\.Baz`)},
		// func (_m *Foo) GetBaz() (*foo.InternalBaz, error)
		{false, regexp.MustCompile(`func \([^\)]+\) GetBaz\(\) \(\*foo\.InternalBaz`)},
	})
}

func (s *GeneratorSuite) TestReplaceTypePackageMultiplePrologueGo123() {
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
	cfg := GeneratorConfig{InPackage: false}

	s.checkGenerationRegexWithConfig("example_project/replace_type/rt.go", "RType", cfg, []regexpExpected{
		// func (_m *RType) Replace1(f rt1.RType1)
		{true, regexp.MustCompile(`func \([^\)]+\) Replace1\(f rt1\.RType1`)},
		// func (_m *RType) Replace2(f rt2.RType2)
		{true, regexp.MustCompile(`func \([^\)]+\) Replace2\(f rt2\.RType2`)},
	})
}
