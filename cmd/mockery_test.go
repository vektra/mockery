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

func TestRunLegacyNoConfig(t *testing.T) {
	tmpDir := t.TempDir()

	mockPath := pathlib.NewPath(tmpDir).Join("Foo.go")
	codePath := pathlib.NewPath(tmpDir).Join("foo.go")
	codePath.WriteFile([]byte(`
package test

type Foo interface {
	Get(str string) string
}`))

	v := viper.New()
	v.Set("log-level", "debug")
	v.Set("outpkg", "foobar")
	v.Set("name", "Foo")
	v.Set("output", tmpDir)
	v.Set("disable-config-search", true)
	os.Chdir(tmpDir)

	initConfig(nil, v, nil)
	app, err := GetRootAppFromViper(v)
	require.NoError(t, err)
	require.NoError(t, app.Run())

	exists, err := mockPath.Exists()
	require.NoError(t, err)
	assert.True(t, exists)
}