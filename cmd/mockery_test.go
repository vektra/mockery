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
	cmd := NewRootCmd()
	assert.Equal(t, "mockery", cmd.Name())
}

func Test_initConfig(t *testing.T) {
	tests := []struct {
		name       string
		base_path  string
		configPath string
	}{
		{
			name:       "test config at base directory",
			base_path:  "1/2/3/4",
			configPath: "1/2/3/4/.mockery.yaml",
		},
		{
			name:       "test config at upper directory",
			base_path:  "1/2/3/4",
			configPath: "1/.mockery.yaml",
		},
		{
			name:      "no config file found",
			base_path: "1/2/3/4",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := pathlib.NewPath(t.TempDir())
			baseDir := tmpDir.Join(strings.Split(tt.base_path, "/")...)
			require.NoError(t, baseDir.MkdirAll())

			configPath := pathlib.NewPath("")
			if tt.configPath != "" {
				configPath = tmpDir.Join(strings.Split(tt.configPath, "/")...)
				configPath.WriteFile([]byte("all: True"))
			}

			viperObj := viper.New()

			initConfig(baseDir, viperObj, nil)

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

	v := viper.New()
	initConfig(nil, v, configPath)
	app, err := GetRootAppFromViper(v)
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

	v := newViper(tmpDir)
	initConfig(nil, v, configPath)
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

	v := newViper(tmpDir)
	initConfig(nil, v, configPath)
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
	//config := fmt.Sprintf(configFmt, tmpDir)
	configPath := pathlib.NewPath(tmpDir).Join("config.yaml")
	require.NoError(t, configPath.WriteFile([]byte(config)))

	goModPath := pathlib.NewPath(tmpDir).Join("go.mod")
	err := goModPath.WriteFile([]byte(`
module github.com/testuser/testpackage
                                                                                                                                                                                 
go 1.20`))
	require.NoError(t, err)

	interfacePath := pathlib.NewPath(tmpDir).Join("internal", "foopkg", "interface.go")
	require.NoError(t, interfacePath.Parent().MkdirAll())
	interfacePath.WriteFile([]byte(`
package foopkg
																																												
type FooInterface interface {
		Foo()
		Bar()
}`))

	mockPath := pathlib.NewPath(tmpDir).Join(
		"mocks",
		"github.com",
		"testuser",
		"testpackage",
		"internal",
		"foopkg",
		"mock_FooInterface.go")

	os.Chdir(tmpDir)

	v := viper.New()
	initConfig(nil, v, configPath)
	app, err := GetRootAppFromViper(v)
	require.NoError(t, err)
	require.NoError(t, app.Run())

	exists, err := mockPath.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRunLegacyNoConfig(t *testing.T) {
	tmpDir := pathlib.NewPath(t.TempDir())

	mockPath := tmpDir.Join("Foo.go")
	codePath := tmpDir.Join("foo.go")
	codePath.WriteFile([]byte(`
package test

type Foo interface {
	Get(str string) string
}`))

	v := viper.New()
	v.Set("log-level", "debug")
	v.Set("outpkg", "foobar")
	v.Set("name", "Foo")
	v.Set("output", tmpDir.String())
	v.Set("disable-config-search", true)
	os.Chdir(tmpDir.String())

	initConfig(nil, v, nil)
	app, err := GetRootAppFromViper(v)
	require.NoError(t, err)
	require.NoError(t, app.Run())

	exists, err := mockPath.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRunLegacyNoConfigDirSet(t *testing.T) {
	tmpDir := pathlib.NewPath(t.TempDir())

	subdir := tmpDir.Join("subdir")
	require.NoError(t, subdir.MkdirAll())

	mockPath := subdir.Join("Foo.go")
	codePath := subdir.Join("foo.go")

	err := codePath.WriteFile([]byte(`
package test

type Foo interface {
	Get(str string) string
}`))
	require.NoError(t, err, "failed to write go file")

	v := viper.New()
	v.Set("log-level", "debug")
	v.Set("outpkg", "foobar")
	v.Set("name", "Foo")
	v.Set("output", subdir.String())
	v.Set("disable-config-search", true)
	v.Set("dir", subdir.String())
	v.Set("recursive", true)
	os.Chdir(tmpDir.String())

	initConfig(nil, v, nil)
	app, err := GetRootAppFromViper(v)
	require.NoError(t, err)
	require.NoError(t, app.Run())

	exists, err := mockPath.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}
