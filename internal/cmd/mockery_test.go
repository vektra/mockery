package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCmd(t *testing.T) {
	cmd, err := NewRootCmd()
	assert.NoError(t, err)
	assert.Equal(t, "mockery", cmd.Name())
}

func Test_initConfig(t *testing.T) {
	tests := []struct {
		name       string
		basePath   string
		configPath string
	}{
		{
			name:       "test config at base directory",
			basePath:   "1/2/3/4",
			configPath: "1/2/3/4/.mockery.yaml",
		},
		{
			name:       "test config at upper directory",
			basePath:   "1/2/3/4",
			configPath: "1/.mockery.yaml",
		},
		{
			name:     "no config file found",
			basePath: "1/2/3/4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := pathlib.NewPath(t.TempDir())
			baseDir := tmpDir.Join(strings.Split(tt.basePath, "/")...)
			require.NoError(t, baseDir.MkdirAll())

			configPath := pathlib.NewPath("")
			if tt.configPath != "" {
				configPath = tmpDir.Join(strings.Split(tt.configPath, "/")...)
				require.NoError(t, configPath.WriteFile([]byte("all: True")))
			}

			viperObj, err := getConfig(baseDir, nil)
			require.NoError(t, err)

			assert.Equal(t, configPath.String(), viperObj.ConfigFileUsed())
		})
	}
}

type Writer interface {
	Foo()
}

func TestRunLegacyGenerationNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	config := `
name: Foo
`
	configPath := pathlib.NewPath(tmpDir).Join("config.yaml")
	require.NoError(t, configPath.WriteFile([]byte(config)))

	viperObj, err := getConfig(nil, configPath)
	require.NoError(t, err)
	app, err := GetRootAppFromViper(viperObj)
	require.NoError(t, err)
	assert.Error(t, app.Run())
}

func newViper(tmpDir string) *viper.Viper {
	v := viper.New()
	v.Set("dir", tmpDir)
	return v
}

func TestRunPackagesGenerationGlobalDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configFmt := `
log-level: info
filename: "hello_{{.InterfaceName}}.go"
packages:
  io:
    config:
      outpkg: mock_io
      dir: %s
    interfaces:
      Writer:`
	config := fmt.Sprintf(configFmt, tmpDir)
	configPath := pathlib.NewPath(tmpDir).Join("config.yaml")
	require.NoError(t, configPath.WriteFile([]byte(config)))
	mockPath := pathlib.NewPath(tmpDir).Join("hello_Writer.go")

	v, err := getConfig(nil, configPath)
	require.NoError(t, err)

	app, err := GetRootAppFromViper(v)
	require.NoError(t, err)
	require.NoError(t, app.Run())

	exists, err := mockPath.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRunPackagesGeneration(t *testing.T) {
	tmpDir := t.TempDir()
	configFmt := `
with-expecter: true
log-level: info
packages:
  io:
    config:
      outpkg: mock_io
      dir: %s
    interfaces:
      Writer:`
	config := fmt.Sprintf(configFmt, tmpDir)
	configPath := pathlib.NewPath(tmpDir).Join("config.yaml")
	require.NoError(t, configPath.WriteFile([]byte(config)))
	mockPath := pathlib.NewPath(tmpDir).Join("mock_Writer.go")

	v, err := getConfig(nil, configPath)
	require.NoError(t, err)
	app, err := GetRootAppFromViper(v)
	require.NoError(t, err)
	require.NoError(t, app.Run())

	exists, err := mockPath.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestIssue565(t *testing.T) {
	// An issue was posed in https://github.com/vektra/mockery/issues/565
	// where mockery wasn't entering the `packages` config section. I think
	// this is some kind of bug with viper. We should instead parse the yaml
	// directly instead of relying on the struct unmarshalling from viper,
	// which is kind of buggy.
	tmpDir := t.TempDir()
	config := `
with-expecter: True
inpackage: True
testonly: True
log-level: debug
packages:
  github.com/testuser/testpackage/internal/foopkg:
    interfaces:
      FooInterface:
`
	configPath := pathlib.NewPath(tmpDir).Join("config.yaml")
	require.NoError(t, configPath.WriteFile([]byte(config)))

	goModPath := pathlib.NewPath(tmpDir).Join("go.mod")
	err := goModPath.WriteFile([]byte(`
module github.com/testuser/testpackage

go 1.20`))
	require.NoError(t, err)

	interfacePath := pathlib.NewPath(tmpDir).Join("internal", "foopkg", "interface.go")
	require.NoError(t, interfacePath.Parent().MkdirAll())
	require.NoError(t, interfacePath.WriteFile([]byte(`
package foopkg

type FooInterface interface {
		Foo()
		Bar()
}`)))

	mockPath := pathlib.NewPath(tmpDir).Join(
		"mocks",
		"github.com",
		"testuser",
		"testpackage",
		"internal",
		"foopkg",
		"mock_FooInterface.go")

	require.NoError(t, os.Chdir(tmpDir))

	v, err := getConfig(nil, configPath)
	require.NoError(t, err)
	app, err := GetRootAppFromViper(v)
	require.NoError(t, err)
	require.NoError(t, app.Run())

	exists, err := mockPath.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}