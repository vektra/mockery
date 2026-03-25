package config

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/types"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/tools/go/packages"
)

func TestNewRootConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr error
	}{
		{
			name: "unrecognized parameter",
			config: `
packages:
  github.com/foo/bar:
    config:
      unknown: param
`,
			wantErr: fmt.Errorf("'packages[github.com/foo/bar].config' has invalid keys: unknown"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := path.Join(t.TempDir(), "config.yaml")
			require.NoError(t, os.WriteFile(configFile, []byte(tt.config), 0o600))

			flags := pflag.NewFlagSet("test", pflag.ExitOnError)
			flags.String("config", "", "")

			require.NoError(t, flags.Parse([]string{"--config", configFile}))

			_, _, err := NewRootConfig(context.Background(), flags)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				var original error
				cursor := err
				for cursor != nil {
					original = cursor
					cursor = errors.Unwrap(cursor)
				}
				assert.Equal(t, tt.wantErr.Error(), original.Error())
			}
		})
	}
}

func TestNewRootConfigUnknownEnvVar(t *testing.T) {
	t.Setenv("MOCKERY_UNKNOWN", "foo")
	configFile := path.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(`
packages:
  github.com/vektra/mockery/v3:
`), 0o600))

	flags := pflag.NewFlagSet("test", pflag.ExitOnError)
	flags.String("config", "", "")

	require.NoError(t, flags.Parse([]string{"--config", configFile}))
	_, _, err := NewRootConfig(context.Background(), flags)
	assert.NoError(t, err)
}

func TestNewRootConfigDefaultFormatterOptions(t *testing.T) {
	configFile := path.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configFile, []byte(`
formatter: goimports
`), 0o600))

	flags := pflag.NewFlagSet("test", pflag.ExitOnError)
	flags.String("config", "", "")

	require.NoError(t, flags.Parse([]string{"--config", configFile}))
	cfg, _, err := NewRootConfig(context.Background(), flags)
	require.NoError(t, err)
	require.NotNil(t, cfg.FormatterOptions.GoImports)
	assert.Equal(t, "goimports", *cfg.Formatter)
	assert.False(t, *cfg.FormatterOptions.GoImports.AllErrors)
	assert.True(t, *cfg.FormatterOptions.GoImports.Comments)
	assert.True(t, *cfg.FormatterOptions.GoImports.FormatOnly)
	assert.Equal(t, "", *cfg.FormatterOptions.GoImports.LocalPrefix)
	assert.True(t, *cfg.FormatterOptions.GoImports.TabIndent)
	assert.Equal(t, 8, *cfg.FormatterOptions.GoImports.TabWidth)
}

func TestExtractConfigFromDirectiveComments(t *testing.T) {
	configs := []struct {
		name         string
		commentLines []string
		expected     *Config
		expectError  bool
	}{
		{
			name: "no directive comments",
			commentLines: []string{
				"// This is a regular comment.",
				"// Another regular comment.",
			},
			expected:    nil,
			expectError: false,
		},
		{
			name: "regular comments are not directive comments",
			commentLines: []string{
				"// Directive comments *must* shouldn't have spaces after the slashes.",
				"// mockery:structname: MyMock",
			},
			expected:    nil,
			expectError: false,
		},
		{
			name: "valid single-line directive comment",
			commentLines: []string{
				"//mockery:structname: MyMock",
			},
			expected: &Config{
				StructName: ptr("MyMock"),
			},
		},
		{
			name: "valid multi-line directive comments",
			commentLines: []string{
				"// Some initial comment.",
				"//mockery:structname: MyMock",
				"//mockery:filename: my_mock.go",
				"// Some trailing comment.",
			},
			expected: &Config{
				StructName: ptr("MyMock"),
				FileName:   ptr("my_mock.go"),
			},
			expectError: false,
		},
		{
			name: "invalid directive comment format",
			commentLines: []string{
				"//mockery:structname MyMock", // Missing ':'
			},
			expected:    nil,
			expectError: true,
		},
		{
			name: "unsupported configuration key are ignored",
			commentLines: []string{
				"//mockery:unknown_key: value",
			},
			expected:    &Config{},
			expectError: false,
		},
		{
			name: "mixed valid and invalid directive comments",
			commentLines: []string{
				"//mockery:structname: MyMock",
				"//mockery:invalid_format", // Invalid
				"//mockery:filename: my_mock.go",
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tt := range configs {
		t.Run(tt.name, func(t *testing.T) {
			comments := make([]*ast.Comment, len(tt.commentLines))
			for i, line := range tt.commentLines {
				comments[i] = &ast.Comment{Text: line}
			}

			result, err := ExtractDirectiveConfig(context.Background(), &ast.GenDecl{
				Doc: &ast.CommentGroup{
					List: comments,
				},
			})
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func ptr[T any](s T) *T {
	return &s
}

func TestBuildTemplateData(t *testing.T) {
	ctx := context.Background()

	configFile := "/config/config.yaml"
	structName := "MyStruct"
	templateName := "my-template"

	c := &Config{
		ConfigFile: &configFile,
		StructName: &structName,
		Template:   &templateName,
	}

	pkg := &packages.Package{
		Types: types.NewPackage("github.com/test/pkg", "mypkg"),
	}

	wd, err := os.Getwd()
	require.NoError(t, err)
	// normalize the working directory path to use forward slashes for consistency across platforms.
	wd = filepath.ToSlash(wd)

	ifacePath := filepath.Join(wd, "internal/foo/bar.go")

	data, err := c.buildTemplateData(
		ctx,
		ifacePath,
		"API",
		pkg,
	)
	require.NoError(t, err)

	assert.Equal(t, "/config", data.ConfigDir)
	assert.Equal(t, fmt.Sprintf("%s/internal/foo", wd), data.InterfaceDir)
	assert.Equal(t, "internal/foo", data.InterfaceDirRelative)
	assert.Equal(t, fmt.Sprintf("%s/internal/foo/bar.go", wd), data.InterfaceFile)
	assert.Equal(t, "API", data.InterfaceName)
	assert.Equal(t, "Mock", data.Mock)
	assert.Equal(t, "MyStruct", data.StructName)
	assert.Equal(t, "mypkg", data.SrcPackageName)
	assert.Equal(t, "github.com/test/pkg", data.SrcPackagePath)
	assert.Equal(t, "my-template", data.Template)
}
