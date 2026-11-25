package cmd

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartCPUProfile(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "cpu.prof")

	stop, err := startCPUProfile(profilePath)
	require.NoError(t, err)
	require.NoError(t, stop())

	info, err := os.Stat(profilePath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestStartCPUProfileBadPath(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "does-not-exist", "cpu.prof")

	_, err := startCPUProfile(profilePath)
	assert.Error(t, err)
}

func TestStartMemProfile(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "mem.prof")

	stop, err := startMemProfile(profilePath)
	require.NoError(t, err)
	require.NoError(t, stop())

	info, err := os.Stat(profilePath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestStartMemProfileBadPath(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "does-not-exist", "mem.prof")

	_, err := startMemProfile(profilePath)
	assert.Error(t, err)
}

func TestRootAppRunConfigurableConstructorName(t *testing.T) {
	outDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outDir, "go.mod"), []byte("module example.com/testout\n\ngo 1.24\n"), 0o600))

	configFile := filepath.Join(outDir, ".mockery.yml")
	configYAML := fmt.Sprintf(`
template: testify
formatter: goimports
force-file-write: true
dir: %q
filename: mocks_out.go
pkgname: testout
constructor-name: "Make{{.InterfaceName}}"
packages:
  github.com/vektra/mockery/v3/internal/fixtures:
    interfaces:
      Requester:
        config:
          constructor-name: 'Custom{{.StructName | trimPrefix "Mock"}}'
      ConstructorTypeParams:
      Blank:
        config:
          constructor-name: ""
  github.com/vektra/mockery/v3/internal/fixtures/unexported:
    config:
      filename: mocks_unexported_out.go
      constructor-name: "newPkg{{.InterfaceName | firstUpper}}"
    interfaces:
      foo:
`, filepath.ToSlash(outDir))
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0o600))

	flags := pflag.NewFlagSet("test", pflag.ExitOnError)
	flags.String("config", "", "")
	require.NoError(t, flags.Parse([]string{"--config", configFile}))

	app, err := GetRootApp(t.Context(), flags)
	require.NoError(t, err)
	require.NoError(t, app.Run())

	generatedBytes, err := os.ReadFile(filepath.Join(outDir, "mocks_out.go"))
	require.NoError(t, err)
	generated := string(generatedBytes)

	_, err = parser.ParseFile(token.NewFileSet(), "mocks_out.go", generatedBytes, parser.AllErrors)
	require.NoError(t, err)

	// Per-interface templated override using resolved StructName.
	assert.Contains(t, generated, "// CustomRequester creates a new instance of MockRequester.")
	assert.Contains(t, generated, "func CustomRequester(t interface {")

	// Top-level templated constructor-name applied to a generic interface.
	assert.Contains(t, generated, "// MakeConstructorTypeParams creates a new instance of MockConstructorTypeParams.")
	assert.Contains(t, generated, "func MakeConstructorTypeParams[helper any, ok any](t interface {")

	// Empty interface-level constructor-name falls back to default New<StructName>.
	assert.Contains(t, generated, "// NewMockBlank creates a new instance of MockBlank.")
	assert.Contains(t, generated, "func NewMockBlank(t interface {")

	// Package-level templated override applied to an unexported interface.
	unexportedBytes, err := os.ReadFile(filepath.Join(outDir, "mocks_unexported_out.go"))
	require.NoError(t, err)
	unexportedGenerated := string(unexportedBytes)

	_, err = parser.ParseFile(token.NewFileSet(), "mocks_unexported_out.go", unexportedBytes, parser.AllErrors)
	require.NoError(t, err)
	assert.Contains(t, unexportedGenerated, "// newPkgFoo creates a new instance of mockfoo.")
	assert.Contains(t, unexportedGenerated, "func newPkgFoo(t interface {")
}

func TestRootAppRunInvalidConstructorNameTemplate(t *testing.T) {
	outDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outDir, "go.mod"), []byte("module example.com/testout\n\ngo 1.24\n"), 0o600))

	configFile := filepath.Join(outDir, ".mockery.yml")
	configYAML := fmt.Sprintf(`
template: testify
dir: %q
filename: mocks_out.go
pkgname: testout
constructor-name: "New{{.DoesNotExist}}"
packages:
  github.com/vektra/mockery/v3/internal/fixtures:
    interfaces:
      Requester:
`, filepath.ToSlash(outDir))
	require.NoError(t, os.WriteFile(configFile, []byte(configYAML), 0o600))

	flags := pflag.NewFlagSet("test", pflag.ExitOnError)
	flags.String("config", "", "")
	require.NoError(t, flags.Parse([]string{"--config", configFile}))

	app, err := GetRootApp(t.Context(), flags)
	require.NoError(t, err)
	err = app.Run()
	require.ErrorContains(t, err, "failed to execute constructor-name template")
}
