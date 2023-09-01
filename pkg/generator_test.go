package pkg

import (
	"bufio"
	"context"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	mocks "github.com/vektra/mockery/v2/mocks/github.com/vektra/mockery/v2/pkg/fixtures"
)

const pkg = "test"

type GeneratorSuite struct {
	suite.Suite
	parser *Parser
	ctx    context.Context
}

func (s *GeneratorSuite) SetupTest() {
	s.parser = NewParser(nil)
	s.ctx = context.Background()
}

func (s *GeneratorSuite) getInterfaceFromFile(interfacePath, interfaceName string) *Interface {
	if !filepath.IsAbs(interfacePath) {
		interfacePath = getFixturePath(interfacePath)
	}
	s.Require().NoError(
		s.parser.Parse(s.ctx, interfacePath),
	)

	s.Require().NoError(
		s.parser.Load(),
	)

	iface, err := s.parser.Find(interfaceName)
	s.Require().NoError(err)
	s.Require().NotNil(iface)
	return iface
}

func (s *GeneratorSuite) getGeneratorWithConfig(
	filepath, interfaceName string, cfg GeneratorConfig,
) *Generator {
	return NewGenerator(s.ctx, cfg, s.getInterfaceFromFile(filepath, interfaceName), pkg)
}

func (s *GeneratorSuite) checkGenerationWithConfig(
	filepath, interfaceName string, cfg GeneratorConfig, expected string,
) *Generator {
	generator := s.getGeneratorWithConfig(filepath, interfaceName, cfg)
	err := generator.Generate(s.ctx)
	s.Require().NoError(err)
	if err != nil {
		return generator
	}

	// Mirror the formatting done by normally done by golang.org/x/tools/imports in Generator.Write.
	//
	// While we could possibly reuse Generator.Write here in addition to Generator.Generate,
	// it would require changing Write's signature to accept custom options, specifically to
	// allow the fragments in preexisting cases. It's assumed that this approximation,
	// just formatting the source, is sufficient for the needs of the current test styles.
	var actual []byte
	actual, fmtErr := format.Source(generator.buf.Bytes())
	s.Require().NoError(fmtErr)

	// Compare lines for easier debugging via testify's slice diff output
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(string(actual), "\n")

	if expected != "" {
		s.Equal(
			expectedLines, actualLines,
			"The generator produced unexpected output.",
		)
	}
	return generator
}

type regexpExpected struct {
	shouldMatch bool
	re          *regexp.Regexp
}

func (s *GeneratorSuite) checkGenerationRegexWithConfig(
	filepath, interfaceName string, cfg GeneratorConfig, expected []regexpExpected,
) *Generator {
	generator := s.getGeneratorWithConfig(filepath, interfaceName, cfg)
	err := generator.Generate(s.ctx)
	s.Require().NoError(err)
	if err != nil {
		return generator
	}
	// Mirror the formatting done by normally done by golang.org/x/tools/imports in Generator.Write.
	//
	// While we could possibly reuse Generator.Write here in addition to Generator.Generate,
	// it would require changing Write's signature to accept custom options, specifically to
	// allow the fragments in preexisting cases. It's assumed that this approximation,
	// just formatting the source, is sufficient for the needs of the current test styles.
	var actual []byte
	actual, fmtErr := format.Source(generator.buf.Bytes())
	s.Require().NoError(fmtErr)

	for _, re := range expected {
		s.Equalf(re.shouldMatch, re.re.Match(actual), "match '%s' should be %t", re.re.String(), re.shouldMatch)
	}

	return generator
}

func (s *GeneratorSuite) getGenerator(
	filepath, interfaceName string, inPackage bool, structName string,
) *Generator {
	return s.getGeneratorWithConfig(filepath, interfaceName, GeneratorConfig{
		StructName:     structName,
		InPackage:      inPackage,
		UnrollVariadic: true,
	})
}

func (s *GeneratorSuite) checkGeneration(filepath, interfaceName string, inPackage bool, structName string, expected string) *Generator {
	cfg := GeneratorConfig{
		StructName:     structName,
		InPackage:      inPackage,
		UnrollVariadic: true,
	}
	return s.checkGenerationWithConfig(filepath, interfaceName, cfg, expected)
}

func (s *GeneratorSuite) checkPrologueGeneration(
	generator *Generator, expected string,
) {
	generator.GeneratePrologue(ctx, "mocks")
	s.Equal(
		expected, generator.buf.String(),
		"The generator produced an unexpected prologue.",
	)
}

func (s *GeneratorSuite) TestGenerator() {
	s.checkGeneration(testFile, "Requester", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorRequesterWithExpecter() {
	cfg := GeneratorConfig{
		WithExpecter:   true,
		UnrollVariadic: false,
	}
	s.checkGenerationWithConfig(testFile, "Requester", cfg, "")
}

func (s *GeneratorSuite) TestGeneratorExpecterComplete() {
	cfg := GeneratorConfig{
		StructName:     "Expecter",
		WithExpecter:   true,
		UnrollVariadic: true,
	}
	s.checkGenerationWithConfig("expecter.go", "Expecter", cfg, "")
}

func (s *GeneratorSuite) TestGeneratorExpecterWithRolledVariadic() {
	expectedBytes, err := os.ReadFile(getMocksPath("ExpecterAndRolledVariadic.go"))
	s.Require().NoError(err)
	expected := string(expectedBytes)
	expected = expected[strings.Index(expected, "// ExpecterAndRolledVariadic is"):]
	generator := NewGenerator(
		s.ctx, GeneratorConfig{
			StructName:     "ExpecterAndRolledVariadic",
			WithExpecter:   true,
			UnrollVariadic: false,
		}, s.getInterfaceFromFile("expecter.go", "Expecter"), pkg,
	)
	s.Require().NoError(generator.Generate(s.ctx))

	var actual []byte
	actual, fmtErr := format.Source(generator.buf.Bytes())
	s.Require().NoError(fmtErr)

	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(string(actual), "\n")

	s.Require().Equal(
		expectedLines, actualLines,
		"The generator produced unexpected output.",
	)
}

func (s *GeneratorSuite) TestGeneratorFunction() {
	s.checkGeneration("function.go", "SendFunc", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorSingleReturn() {
	s.checkGeneration(testFile2, "Requester2", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorNoArguments() {
	s.checkGeneration("requester3.go", "Requester3", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorNoNothing() {
	s.checkGeneration("requester4.go", "Requester4", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorUnexported() {
	s.checkGeneration("requester_unexported.go", "requester_unexported", true, "", "")
}

func (s *GeneratorSuite) TestGeneratorPrologue() {
	generator := s.getGenerator(testFile, "Requester", false, "")
	expected := `package mocks

import mock "github.com/stretchr/testify/mock"
import test "github.com/vektra/mockery/v2/pkg/fixtures"

`
	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestGeneratorPrologueWithImports() {
	generator := s.getGenerator("requester_ns.go", "RequesterNS", false, "")
	expected := `package mocks

import http "net/http"
import mock "github.com/stretchr/testify/mock"
import test "github.com/vektra/mockery/v2/pkg/fixtures"

`
	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestGeneratorPrologueWithMultipleImportsSameName() {
	generator := s.getGenerator("same_name_imports.go", "Example", false, "")

	expected := `package mocks

import fixtureshttp "github.com/vektra/mockery/v2/pkg/fixtures/http"
import http "net/http"
import mock "github.com/stretchr/testify/mock"
import test "github.com/vektra/mockery/v2/pkg/fixtures"

`
	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestGeneratorPrologueNote() {
	generator := s.getGenerator(testFile, "Requester", false, "")
	generator.GeneratePrologueNote("A\\nB")

	expected := `// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

// A
// B

`

	s.Equal(expected, generator.buf.String())
}

func (s *GeneratorSuite) TestGeneratorBoilerplate() {
	generator := s.getGenerator(testFile, "Requester", false, "")
	generator.GenerateBoilerplate("/*\n    BOILERPLATE\n*/\n")

	expected := `/*
    BOILERPLATE
*/

`

	s.Equal(expected, generator.buf.String())
}

func (s *GeneratorSuite) TestGeneratorPrologueNoteNoVersionString() {
	generator := s.getGenerator(testFile, "Requester", false, "")
	generator.config.DisableVersionString = true
	generator.GeneratePrologueNote("A\\nB")

	expected := `// Code generated by mockery. DO NOT EDIT.

// A
// B

`

	s.Equal(expected, generator.buf.String())
}

func (s *GeneratorSuite) TestVersionOnCorrectLine() {
	gen := s.getGenerator(testFile, "Requester", false, "")

	// Run everything that is ran by the GeneratorVisitor
	gen.GeneratePrologueNote("A\\nB")
	gen.GeneratePrologue(s.ctx, pkg)
	err := gen.Generate(s.ctx)

	s.Require().NoError(err)
	scan := bufio.NewScanner(&gen.buf)
	s.Contains("Code generated by", scan.Text())
}

func (s *GeneratorSuite) TestGeneratorChecksInterfacesForNilable() {
	s.checkGeneration("requester_iface.go", "RequesterIface", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorPointers() {
	s.checkGeneration("requester_ptr.go", "RequesterPtr", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorSlice() {
	s.checkGeneration("requester_slice.go", "RequesterSlice", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorArrayLiteralLen() {
	s.checkGeneration("requester_array.go", "RequesterArray", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorNamespacedTypes() {
	s.checkGeneration("requester_ns.go", "RequesterNS", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorWhereArgumentNameConflictsWithImport() {
	s.checkGeneration("requester_arg_same_as_import.go", "RequesterArgSameAsImport", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorWhereArgumentNameConflictsWithNamedImport() {
	s.checkGeneration("requester_arg_same_as_named_import.go", "RequesterArgSameAsNamedImport", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorWhereArgumentNameConflictsWithPackage() {
	s.checkGeneration("requester_arg_same_as_pkg.go", "RequesterArgSameAsPkg", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorHavingNoNamesOnArguments() {
	s.checkGeneration("custom_error.go", "KeyManager", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorElidedType() {
	s.checkGeneration("requester_elided.go", "RequesterElided", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorReturnElidedType() {
	cfg := GeneratorConfig{
		WithExpecter: true,
	}

	s.checkGenerationWithConfig("requester_ret_elided.go", "RequesterReturnElided", cfg, "")
}

func (s *GeneratorSuite) TestGeneratorVariadicArgs() {
	expectedBytes, err := os.ReadFile(getMocksPath("RequesterVariadic.go"))
	s.Require().NoError(err)
	expected := string(expectedBytes)
	expected = expected[strings.Index(expected, "// RequesterVariadic is"):]
	s.checkGeneration("requester_variadic.go", "RequesterVariadic", false, "", expected)
}

func (s *GeneratorSuite) TestGeneratorVariadicArgsAsOneArg() {
	expectedBytes, err := os.ReadFile(getMocksPath("RequesterVariadicOneArgument.go"))
	s.Require().NoError(err)
	expected := string(expectedBytes)
	expected = expected[strings.Index(expected, "// RequesterVariadicOneArgument is"):]
	generator := NewGenerator(
		s.ctx, GeneratorConfig{
			StructName:     "RequesterVariadicOneArgument",
			InPackage:      true,
			UnrollVariadic: false,
		}, s.getInterfaceFromFile("requester_variadic.go", "RequesterVariadic"), pkg,
	)
	s.Require().NoError(generator.Generate(s.ctx))

	var actual []byte
	actual, fmtErr := format.Source(generator.buf.Bytes())
	s.Require().NoError(fmtErr)

	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(string(actual), "\n")

	s.Equal(
		expectedLines, actualLines,
		"The generator produced unexpected output.",
	)
}

func TestRequesterVariadicOneArgument(t *testing.T) {
	t.Run("Get \"1\", \"2\", \"3\"", func(t *testing.T) {
		m := mocks.RequesterVariadicOneArgument{}
		args := []string{"1", "2", "3"}
		m.On("Get", args).Return(true).Once()
		res := m.Get(args...)
		assert.True(t, res)
		m.AssertExpectations(t)
	})

	t.Run("Get mock.Anything", func(t *testing.T) {
		m := mocks.RequesterVariadicOneArgument{}
		args := []string{"1", "2", "3"}
		m.On("Get", mock.Anything).Return(true).Once()
		res := m.Get(args...)
		assert.True(t, res)
		m.AssertExpectations(t)
	})

	t.Run("Get no arguments", func(t *testing.T) {
		m := mocks.RequesterVariadicOneArgument{}
		m.On("Get", []string(nil)).Return(true).Once()
		res := m.Get()
		assert.True(t, res)
		m.AssertExpectations(t)
	})

	t.Run("MultiWriteToFile strings builders", func(t *testing.T) {
		m := mocks.RequesterVariadicOneArgument{}
		args := []io.Writer{&strings.Builder{}, &strings.Builder{}}
		expected := "res"
		filename := "testFilename"
		m.On("MultiWriteToFile", filename, args).Return(expected).Once()
		res := m.MultiWriteToFile(filename, args...)
		assert.Equal(t, expected, res)
		m.AssertExpectations(t)
	})

	t.Run("MultiWriteToFile mock.Anything", func(t *testing.T) {
		m := mocks.RequesterVariadicOneArgument{}
		args := []io.Writer{&strings.Builder{}, &strings.Builder{}}
		expected := "res"
		filename := "testFilename"
		m.On("MultiWriteToFile", filename, mock.Anything).Return(expected).Once()
		res := m.MultiWriteToFile(filename, args...)
		assert.Equal(t, expected, res)
		m.AssertExpectations(t)
	})

	t.Run("OneInterface \"1\", \"2\", \"3\"", func(t *testing.T) {
		m := mocks.RequesterVariadicOneArgument{}
		args := []interface{}{"1", "2", "3"}
		m.On("OneInterface", args).Return(true).Once()
		res := m.OneInterface(args...)
		assert.True(t, res)
		m.AssertExpectations(t)
	})

	t.Run("OneInterface mock.Anything", func(t *testing.T) {
		m := mocks.RequesterVariadicOneArgument{}
		args := []interface{}{"1", "2", "3"}
		m.On("OneInterface", mock.Anything).Return(true).Once()
		res := m.OneInterface(args...)
		assert.True(t, res)
		m.AssertExpectations(t)
	})

	t.Run("Sprintf strings builders", func(t *testing.T) {
		m := mocks.RequesterVariadicOneArgument{}
		args := []interface{}{&strings.Builder{}, &strings.Builder{}}
		expected := "res"
		filename := "testFilename"
		m.On("Sprintf", filename, args).Return(expected).Once()
		res := m.Sprintf(filename, args...)
		assert.Equal(t, expected, res)
		m.AssertExpectations(t)
	})

	t.Run("Sprintf mock.Anything", func(t *testing.T) {
		m := mocks.RequesterVariadicOneArgument{}
		args := []interface{}{&strings.Builder{}, &strings.Builder{}}
		expected := "res"
		filename := "testFilename"
		m.On("Sprintf", filename, mock.Anything).Return(expected).Once()
		res := m.Sprintf(filename, args...)
		assert.Equal(t, expected, res)
		m.AssertExpectations(t)
	})
}

func (s *GeneratorSuite) TestGeneratorArgumentIsFuncType() {
	s.checkGeneration("func_type.go", "Fooer", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorChanType() {
	s.checkGeneration("async.go", "AsyncProducer", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorFromImport() {
	s.checkGeneration("io_import.go", "MyReader", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorComplexChanFromConsul() {
	s.checkGeneration("consul.go", "ConsulLock", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorForEmptyInterface() {
	s.checkGeneration("empty_interface.go", "Blank", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorArgumentIsMapFunc() {
	s.checkGeneration("map_func.go", "MapFunc", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorForMethodUsingInterface() {
	s.checkGeneration("mock_method_uses_pkg_iface.go", "UsesOtherPkgIface", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorForMethodUsingInterfaceInPackage() {
	s.checkGeneration("mock_method_uses_pkg_iface.go", "UsesOtherPkgIface", true, "", "")
}

func (s *GeneratorSuite) TestGeneratorWithAliasing() {
	s.checkGeneration("same_name_imports.go", "Example", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorWithImportSameAsLocalPackageInpkgNoCycle() {
	iface := s.getInterfaceFromFile("imports_same_as_package.go", "ImportsSameAsPackage")
	pkg := iface.QualifiedName
	gen := NewGenerator(s.ctx, GeneratorConfig{
		InPackage: true,
	}, iface, pkg)
	gen.GeneratePrologue(s.ctx, pkg)
	s.NotContains(gen.buf.String(), `import test "github.com/vektra/mockery/v2/pkg/fixtures/test"`)
}

func (s *GeneratorSuite) TestMapToInterface() {
	s.checkGeneration("map_to_interface.go", "MapToInterface", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorFunctionArgsNamesCollision() {
	s.checkGeneration("func_args_collision.go", "FuncArgsCollision", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorWithImportSameAsLocalPackage() {
	s.checkGeneration("imports_same_as_package.go", "ImportsSameAsPackage", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorWithUnsafePointer() {
	s.checkGeneration("unsafe.go", "UnsafeInterface", false, "", "")
}

func (s *GeneratorSuite) TestPrologueWithImportSameAsLocalPackage() {
	generator := s.getGenerator(
		"imports_same_as_package.go", "ImportsSameAsPackage", false, "",
	)
	expected := `package mocks

import fixtures "` + generator.iface.QualifiedName + `"
import mock "github.com/stretchr/testify/mock"
import test "github.com/vektra/mockery/v2/pkg/fixtures/redefined_type_b"

`
	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestPrologueWithImportFromNestedInterface() {
	generator := s.getGenerator(
		"imports_from_nested_interface.go", "HasConflictingNestedImports", false, "",
	)
	expected := `package mocks

import fixtureshttp "github.com/vektra/mockery/v2/pkg/fixtures/http"
import http "net/http"
import mock "github.com/stretchr/testify/mock"
import test "github.com/vektra/mockery/v2/pkg/fixtures"

`

	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestGeneratorForStructValueReturn() {
	s.checkGeneration("struct_value.go", "A", false, "", "")
}

func (s *GeneratorSuite) TestGeneratorForStructWithTag() {
	// StructTag has back-quote, So can't use raw string literals in this test case.
	var expected string
	expected += "*struct {"
	expected += "FieldC int `json:\"field_c\"`"
	expected += "FieldD int `json:\"field_d\" xml:\"field_d\"`"
	expected += "}"

	gen := s.getGeneratorWithConfig("struct_with_tag.go", "StructWithTag", GeneratorConfig{})
	err := gen.Generate(s.ctx)
	s.Require().NoError(err)

	actual := bufio.NewScanner(&gen.buf).Text()
	s.Contains(expected, actual)
}

func (s *GeneratorSuite) TestStructNameOverride() {
	s.checkGeneration(testFile2, "Requester2", false, "Requester2OverrideName", "")
}

func (s *GeneratorSuite) TestKeepTreeInPackageCombined() {
	type testData struct {
		path     string
		name     string
		expected string
	}

	tests := []testData{
		{path: filepath.Join("example_project", "root.go"), name: "Root", expected: `package example_project

import fixturesexample_project "github.com/vektra/mockery/v2/pkg/fixtures/example_project"
import foo "github.com/vektra/mockery/v2/pkg/fixtures/example_project/foo"
import mock "github.com/stretchr/testify/mock"

`},
		{path: filepath.Join("example_project", "foo", "foo.go"), name: "Foo", expected: `package foo

import example_projectfoo "github.com/vektra/mockery/v2/pkg/fixtures/example_project/foo"
import mock "github.com/stretchr/testify/mock"

`},
	}

	for _, test := range tests {
		generator := NewGenerator(
			s.ctx,
			GeneratorConfig{InPackage: true, KeepTree: true},
			s.getInterfaceFromFile(test.path, test.name),
			pkg,
		)
		s.checkPrologueGeneration(generator, test.expected)
	}
}

func (s *GeneratorSuite) TestInPackagePackageCollision() {
	expected := `package foo

import barfoo "github.com/vektra/mockery/v2/pkg/fixtures/example_project/bar/foo"
import mock "github.com/stretchr/testify/mock"

`
	generator := NewGenerator(
		s.ctx,
		GeneratorConfig{InPackage: true},
		s.getInterfaceFromFile("example_project/foo/pkg_name_same_as_import.go", "PackageNameSameAsImport"),
		pkg,
	)
	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestImportCollideWithStdLib() {
	expected := `package context

import context2 "context"
import mock "github.com/stretchr/testify/mock"

`
	generator := NewGenerator(
		s.ctx,
		GeneratorConfig{InPackage: true},
		s.getInterfaceFromFile("example_project/context/context.go", "CollideWithStdLib"),
		pkg,
	)
	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestReplaceTypePackagePrologue() {
	expected := `package mocks

import baz "github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz"
import mock "github.com/stretchr/testify/mock"

`
	generator := NewGenerator(
		s.ctx,
		GeneratorConfig{InPackage: false, ReplaceType: []string{
			"github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz/internal/foo.InternalBaz=baz:github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz.Baz",
		}},
		s.getInterfaceFromFile("example_project/baz/foo.go", "Foo"),
		pkg,
	)

	s.checkPrologueGeneration(generator, expected)
}

func (s *GeneratorSuite) TestReplaceTypePackage() {
	cfg := GeneratorConfig{InPackage: false, ReplaceType: []string{
		"github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz/internal/foo.InternalBaz=baz:github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz.Baz",
	}}

	s.checkGenerationRegexWithConfig("example_project/baz/foo.go", "Foo", cfg, []regexpExpected{
		// func (_m *Foo) GetBaz() (*baz.Baz, error)
		{true, regexp.MustCompile(`func \([^\)]+\) GetBaz\(\) \(\*baz\.Baz`)},
		// func (_m *Foo) GetBaz() (*foo.InternalBaz, error)
		{false, regexp.MustCompile(`func \([^\)]+\) GetBaz\(\) \(\*foo\.InternalBaz`)},
	})
}

func (s *GeneratorSuite) TestGenericGenerator() {
	s.checkGeneration("generic.go", "RequesterGenerics", false, "", "")
}

func (s *GeneratorSuite) TestGenericExpecterGenerator() {
	cfg := GeneratorConfig{
		StructName:     "RequesterGenerics",
		WithExpecter:   true,
		UnrollVariadic: true,
	}
	s.checkGenerationWithConfig("generic.go", "RequesterGenerics", cfg, "")
}

func (s *GeneratorSuite) TestGenericInpkgGenerator() {
	s.checkGeneration("generic.go", "RequesterGenerics", true, "", "")
}

func TestGeneratorSuite(t *testing.T) {
	generatorSuite := new(GeneratorSuite)
	suite.Run(t, generatorSuite)
}

func TestParseReplaceType(t *testing.T) {
	tests := []struct {
		value    string
		expected replaceType
	}{
		{
			value:    "github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz/internal/foo.InternalBaz",
			expected: replaceType{alias: "", pkg: "github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz/internal/foo", typ: "InternalBaz"},
		},
		{
			value:    "baz:github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz.Baz",
			expected: replaceType{alias: "baz", pkg: "github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz", typ: "Baz"},
		},
		{
			value:    "github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz",
			expected: replaceType{alias: "", pkg: "github.com/vektra/mockery/v2/pkg/fixtures/example_project/baz", typ: ""},
		},
	}

	for _, test := range tests {
		actual := parseReplaceType(test.value)
		assert.Equal(t, test.expected, *actual)
	}
}
