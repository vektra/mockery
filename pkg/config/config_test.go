package config

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/mockery/v2/pkg/logging"
	"gopkg.in/yaml.v3"
)

func TestConfig_GetPackageConfig(t *testing.T) {
	type fields struct {
		All       bool
		BuildTags string
		Case      string
		Packages  map[string]interface{}
	}
	type args struct {
		packageName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Config
		wantErr bool
		repeat  uint
	}{
		{
			name: "no config set on package-level config",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{},
				},
			},
			args: args{
				packageName: "github.com/vektra/mockery/v2/pkg",
			},
			want: &Config{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{},
				},
			},
		},
		{
			name: "package not defined in config",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages:  map[string]any{},
			},
			args: args{
				packageName: "github.com/vektra/mockery/v2/pkg",
			},
			wantErr: true,
			want:    nil,
		},
		{
			name: "config section provided but no values defined",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{},
					},
				},
			},
			args: args{
				packageName: "github.com/vektra/mockery/v2/pkg",
			},
			want: &Config{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{},
				},
			},
		},
		{
			name: "two values overridden in pkg config",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"all":  false,
							"tags": "foobar",
						},
					},
				},
			},
			args: args{
				packageName: "github.com/vektra/mockery/v2/pkg",
			},
			want: &Config{
				All:       false,
				BuildTags: "foobar",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{},
				},
			},
		},
		{
			name: "repeated calls gives same cached result",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"all":  false,
							"tags": "foobar",
						},
					},
				},
			},
			args: args{
				packageName: "github.com/vektra/mockery/v2/pkg",
			},
			want: &Config{
				All:       false,
				BuildTags: "foobar",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{},
				},
			},
			repeat: 2,
		},
		{
			name: "invalid key provided in config",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"huh?": false,
						},
					},
				},
			},
			args: args{
				packageName: "github.com/vektra/mockery/v2/pkg",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				All:       tt.fields.All,
				BuildTags: tt.fields.BuildTags,
				Case:      tt.fields.Case,
				Packages:  tt.fields.Packages,
			}
			c.Config = writeConfigFile(t, c)
			log, err := logging.GetLogger("DEBUG")
			require.NoError(t, err)

			if tt.repeat == 0 {
				tt.repeat = 1
			}

			for i := uint(0); i < tt.repeat; i++ {
				got, err := c.GetPackageConfig(log.WithContext(context.Background()), tt.args.packageName)
				if (err != nil) != tt.wantErr {
					t.Errorf("Config.GetPackageConfig() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				if tt.want == nil {
					assert.Nil(t, got)
				} else {
					assert.Equal(t, tt.want.All, got.All)
					assert.Equal(t, tt.want.BuildTags, got.BuildTags)
					assert.Equal(t, tt.want.Case, got.Case)
				}
			}
		})
	}
}

func TestConfig_GetInterfaceConfig(t *testing.T) {
	type fields struct {
		All       bool
		BuildTags string
		Case      string
		Packages  map[string]interface{}
	}
	type args struct {
		packageName   string
		interfaceName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Config
		wantErr bool
	}{
		{
			name: "no config defined for package",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{},
				},
			},
			args: args{
				packageName:   "github.com/vektra/mockery/v2/pkg",
				interfaceName: "intf",
			},
			want: []*Config{
				{
					All:       true,
					BuildTags: "default_tags",
					Case:      "upper",
				},
			},
		},
		{
			name: "config defined for package",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"all": false,
						},
					},
				},
			},
			args: args{
				packageName:   "github.com/vektra/mockery/v2/pkg",
				interfaceName: "intf",
			},
			want: []*Config{
				{
					All:       false,
					BuildTags: "default_tags",
					Case:      "upper",
				},
			},
		},
		{
			name: "empty interfaces section",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"all": false,
						},
						"interfaces": map[string]any{},
					},
				},
			},
			args: args{
				packageName:   "github.com/vektra/mockery/v2/pkg",
				interfaceName: "intf",
			},
			want: []*Config{
				{
					All:       false,
					BuildTags: "default_tags",
					Case:      "upper",
				},
			},
		},
		{
			name: "interface defined, but not config section",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"all": false,
						},
						"interfaces": map[string]any{
							"intf": map[string]any{},
						},
					},
				},
			},
			args: args{
				packageName:   "github.com/vektra/mockery/v2/pkg",
				interfaceName: "intf",
			},
			want: []*Config{
				{
					All:       false,
					BuildTags: "default_tags",
					Case:      "upper",
				},
			},
		},
		{
			name: "interface defined with empty config section",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"all": false,
						},
						"interfaces": map[string]any{
							"intf": map[string]any{
								"config": map[string]any{},
							},
						},
					},
				},
			},
			args: args{
				packageName:   "github.com/vektra/mockery/v2/pkg",
				interfaceName: "intf",
			},
			want: []*Config{
				{
					All:       false,
					BuildTags: "default_tags",
					Case:      "upper",
				},
			},
		},
		{
			name: "interface defined with non-empty config",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"all": false,
						},
						"interfaces": map[string]any{
							"intf": map[string]any{
								"config": map[string]any{
									"tags": "foobar",
								},
							},
						},
					},
				},
			},
			args: args{
				packageName:   "github.com/vektra/mockery/v2/pkg",
				interfaceName: "intf",
			},
			want: []*Config{
				{
					All:       false,
					BuildTags: "foobar",
					Case:      "upper",
				},
			},
		},
		{
			name: "interface defined with multiple config entries",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"all": false,
						},
						"interfaces": map[string]any{
							"intf": map[string]any{
								"config": map[string]any{
									"tags": "foobar",
								},
								"configs": []map[string]any{
									{
										"all":  true,
										"tags": "bat",
									},
									{
										"case": "lower",
									},
								},
							},
						},
					},
				},
			},
			args: args{
				packageName:   "github.com/vektra/mockery/v2/pkg",
				interfaceName: "intf",
			},
			want: []*Config{
				{
					All:       true,
					BuildTags: "bat",
					Case:      "upper",
				},
				{
					All:       false,
					BuildTags: "foobar",
					Case:      "lower",
				},
			},
		},
		{
			name: "interface defined with invalid key",
			fields: fields{
				All:       true,
				BuildTags: "default_tags",
				Case:      "upper",
				Packages: map[string]any{
					"github.com/vektra/mockery/v2/pkg": map[string]any{
						"config": map[string]any{
							"all": false,
						},
						"interfaces": map[string]any{
							"FooBarBat": map[string]any{
								"config": map[string]any{
									"invalid-key": "foobar",
								},
							},
						},
					},
				},
			},
			args: args{
				packageName:   "github.com/vektra/mockery/v2/pkg",
				interfaceName: "FooBarBat",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				All:       tt.fields.All,
				BuildTags: tt.fields.BuildTags,
				Case:      tt.fields.Case,
				Packages:  tt.fields.Packages,
			}
			configPath := writeConfigFile(t, c)
			c.Config = configPath

			log, err := logging.GetLogger("DEBUG")
			require.NoError(t, err)

			got, err := c.GetInterfaceConfig(log.WithContext(context.Background()), tt.args.packageName, tt.args.interfaceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.GetPackageConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil {
				assert.Nil(t, got)
			} else {
				assert.Equal(t, len(tt.want), len(got))
				for idx, entry := range got {
					if idx >= len(tt.want) {
						break
					}
					assert.Equal(t, tt.want[idx].All, entry.All)
					assert.Equal(t, tt.want[idx].BuildTags, entry.BuildTags)
					assert.Equal(t, tt.want[idx].Case, entry.Case)
				}
			}
		})
	}
}

func writeConfigFile(t *testing.T, c *Config) string {
	configFile := pathlib.NewPath(t.TempDir()).Join("config.yaml")
	var yamlBuffer bytes.Buffer
	encoder := yaml.NewEncoder(&yamlBuffer)
	defer encoder.Close()
	require.NoError(t, encoder.Encode(c))
	require.NoError(t, configFile.WriteFile(yamlBuffer.Bytes()))
	return configFile.String()
}

func TestConfig_GetPackages(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    []string
		wantErr bool
	}{
		{
			name: "empty config",
			yaml: ``,
			want: []string{},
		},
		{
			name:    "packages defined but no value",
			yaml:    `packages:`,
			want:    []string{},
			wantErr: true,
		},
		{
			name: "packages defined with single package",
			yaml: `packages:
  github.com/vektra/mockery/v2/pkg:`,
			want: []string{"github.com/vektra/mockery/v2/pkg"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := pathlib.NewPath(t.TempDir())
			configFile := tmpDir.Join(".mockery.yaml")
			require.NoError(t, configFile.WriteFile([]byte(tt.yaml)))

			c := &Config{
				Config: configFile.String(),
			}
			got, err := c.GetPackages(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.GetPackages() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Config.GetPackages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_ShouldGenerateInterface(t *testing.T) {
	tests := []struct {
		name    string
		c       *Config
		want    bool
		wantErr bool
	}{
		{
			name: "no packages return error",
			c: &Config{
				Packages: map[string]interface{}{},
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "should generate all interfaces",
			c: &Config{
				Packages: map[string]interface{}{
					"some_package": map[string]interface{}{},
				},
				All: true,
			},
			want: true,
		},
		{
			name: "should generate this package",
			c: &Config{
				Packages: map[string]interface{}{
					"some_package": map[string]interface{}{
						"config": map[string]interface{}{
							"all": true,
						},
					},
				},
			},
			want: true,
		},
		{
			name: "should generate this interface",
			c: &Config{
				Packages: map[string]interface{}{
					"some_package": map[string]interface{}{
						"interfaces": map[string]interface{}{
							"SomeInterface": struct{}{},
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Config = writeConfigFile(t, tt.c)

			got, err := tt.c.ShouldGenerateInterface(context.Background(), "some_package", "SomeInterface")
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.ShouldGenerateInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Config.ShouldGenerateInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_ExcludePath(t *testing.T) {
	tests := []struct {
		name string
		file string
		c    *Config
		want bool
	}{
		{
			name: "should not exclude",
			file: "some_foo.go",
			c: &Config{
				Exclude: []string{"foo"},
			},
			want: false,
		},
		{
			name: "should not exclude both",
			file: "some_foo.go",
			c: &Config{
				Exclude: []string{"foo", "bar"},
			},
			want: false,
		},
		{
			name: "should exclude",
			file: "foo/some_foo.go",
			c: &Config{
				Exclude: []string{"foo"},
			},
			want: true,
		},
		{
			name: "should exclude specific file",
			file: "foo/some_foo.go",
			c: &Config{
				Exclude: []string{"foo/some_foo.go"},
			},
			want: true,
		},
		{
			name: "should exclude both paths",
			file: "foo/bar/some_foo.go",
			c: &Config{
				Exclude: []string{"foo", "foo/bar"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.c.Config = writeConfigFile(t, tt.c)

			got := tt.c.ExcludePath(tt.file)
			if got != tt.want {
				t.Errorf("Config.ExcludePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewConfigFromViper(t *testing.T) {
	tests := []struct {
		name    string
		v       func(t *testing.T) *viper.Viper
		yaml    string
		want    *Config
		wantErr bool
	}{
		{
			name: "default dir",
			v: func(t *testing.T) *viper.Viper {
				return viper.New()
			},
			want: &Config{
				Case:   "camel",
				Dir:    ".",
				Output: "./mocks",
			},
		},
		{
			name: "default packages variables",
			yaml: `
packages:
  github.com/vektra/mockery/v2/pkg:
`,
			want: &Config{
				Dir:                  "mocks/{{.PackagePath}}",
				FileName:             "mock_{{.InterfaceName}}.go",
				IncludeAutoGenerated: true,
				MockName:             "Mock{{.InterfaceName}}",
				Outpkg:               "{{.PackageName}}",
				WithExpecter:         true,
				LogLevel:             "info",
			},
		},
		{
			name: "packages filename set at top level",
			yaml: `
dir: barfoo
filename: foobar.go
packages:
  github.com/vektra/mockery/v2/pkg:
`,
			want: &Config{
				Dir:                  "barfoo",
				FileName:             "foobar.go",
				IncludeAutoGenerated: true,
				MockName:             "Mock{{.InterfaceName}}",
				Outpkg:               "{{.PackageName}}",
				WithExpecter:         true,
				LogLevel:             "info",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var viperObj *viper.Viper
			if tt.v == nil {
				viperObj = viper.New()
			} else {
				viperObj = tt.v(t)
			}

			viperObj.SetConfigName((".mockery"))
			if tt.yaml != "" {
				confPath := pathlib.NewPath(t.TempDir()).Join(".mockery.yaml")
				require.NoError(t, confPath.WriteFile([]byte(tt.yaml)))
				viperObj.AddConfigPath(confPath.Parent().String())
				require.NoError(t, viperObj.ReadInConfig())

				tt.want.Config = confPath.String()
			}
			got, err := NewConfigFromViper(viperObj)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfigFromViper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// zero these out as it's an implementation detail we don't
			// are about testing
			got._cfgAsMap = nil
			tt.want._cfgAsMap = nil

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfigFromViper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Initialize(t *testing.T) {
	tests := []struct {
		name       string
		cfgYaml    string
		wantCfgMap string
		wantErr    error
	}{
		{
			name: "package with no go files",
			cfgYaml: `
packages:
  github.com/vektra/mockery/v2/pkg/fixtures/pkg_with_no_files:
    config:
      recursive: True
      all: True`,
			wantErr: ErrNoGoFilesFoundInRoot,
		},
		{
			name: "test with no subpackages present",
			cfgYaml: `
packages:
  github.com/vektra/mockery/v2/pkg/fixtures/example_project/foo:
    config:
      recursive: True
      all: True`,
			wantCfgMap: `packages:
    github.com/vektra/mockery/v2/pkg/fixtures/example_project/foo:
        config:
            all: true
            recursive: true
`,
		},
		{
			name: "test with one subpackage present",
			cfgYaml: `
with-expecter: False
dir: foobar
packages:
  github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2:
    config:
      recursive: True
      with-expecter: True
      all: True`,
			wantCfgMap: `dir: foobar
packages:
    github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2:
        config:
            all: true
            dir: foobar
            recursive: true
            with-expecter: true
    github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2/subpkg3:
        config:
            all: true
            dir: foobar
            recursive: true
            with-expecter: true
with-expecter: false
`,
		},
		{
			name: "test with one subpackage, config already defined",
			cfgYaml: `
with-expecter: False
dir: foobar
packages:
  github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2:
    config:
      recursive: True
      with-expecter: True
      all: True
  github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2/subpkg3:
    config:
      recursive: True
      with-expecter: True
      all: false
      note: note
      dir: barbaz`,
			wantCfgMap: `dir: foobar
packages:
    github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2:
        config:
            all: true
            dir: foobar
            recursive: true
            with-expecter: true
    github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2/subpkg3:
        config:
            all: false
            dir: barbaz
            note: note
            recursive: true
            with-expecter: true
with-expecter: false
`,
		},
		{
			name: "test with one subpackage, config not defined",
			cfgYaml: `
with-expecter: False
dir: foobar
packages:
  github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2:
    config:
      recursive: True
      with-expecter: True
      all: True
  github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2/subpkg3: {}
`,
			wantCfgMap: `dir: foobar
packages:
    github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2:
        config:
            all: true
            dir: foobar
            recursive: true
            with-expecter: true
    github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2/subpkg3:
        config:
            all: true
            dir: foobar
            recursive: true
            with-expecter: true
with-expecter: false
`,
		},
		{
			name: "test with subpackage's interfaces defined",
			cfgYaml: `
with-expecter: False
dir: foobar
packages:
  github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2:
    config:
      recursive: True
      with-expecter: True
      all: True
  github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2/subpkg3:
    interfaces:
      Getter:
        config:
          with-expecter: False`,
			wantCfgMap: `dir: foobar
packages:
    github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2:
        config:
            all: true
            dir: foobar
            recursive: true
            with-expecter: true
    github.com/vektra/mockery/v2/pkg/fixtures/example_project/pkg_with_subpkgs/subpkg2/subpkg3:
        config:
            all: true
            dir: foobar
            recursive: true
            with-expecter: true
        interfaces:
            Getter:
                config:
                    all: true
                    dir: foobar
                    recursive: true
                    with-expecter: false
with-expecter: false
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpdir := pathlib.NewPath(t.TempDir())
			cfg := tmpdir.Join("config.yaml")
			require.NoError(t, cfg.WriteFile([]byte(tt.cfgYaml)))

			c := &Config{
				Config: cfg.String(),
			}
			log, err := logging.GetLogger("TRACE")
			require.NoError(t, err)

			if err := c.Initialize(log.WithContext(context.Background())); !errors.Is(err, tt.wantErr) {
				t.Errorf("Config.Initialize() error = %v, wantErr %v", err, tt.wantErr)
			}

			cfgAsStr, err := yaml.Marshal(c._cfgAsMap)
			require.NoError(t, err)

			if tt.wantCfgMap != "" && !reflect.DeepEqual(string(cfgAsStr), tt.wantCfgMap) {
				t.Errorf(`Config.Initialize resultant config map
got
----
%v

want
------
%v`, string(cfgAsStr), tt.wantCfgMap)
			}

		})
	}
}

func Test_isAutoGenerated(t *testing.T) {
	tests := []struct {
		name    string
		args    func(t *testing.T) *pathlib.Path
		want    bool
		wantErr bool
	}{
		{
			name: "file is autogenerated",
			args: func(t *testing.T) *pathlib.Path {
				path := pathlib.NewPath(t.TempDir()).Join("file.go")
				assert.NoError(t, path.WriteFile([]byte("// Code generated by mockery. DO NOT EDIT.")))
				return path
			},
			want: true,
		},
		{
			name: "file is not autogenerated",
			args: func(t *testing.T) *pathlib.Path {
				path := pathlib.NewPath(t.TempDir()).Join("file.go")
				assert.NoError(t, path.WriteFile([]byte("package foobar")))
				return path
			},
			want: false,
		},
		{
			name: "error when reading file",
			args: func(t *testing.T) *pathlib.Path {
				path := pathlib.NewPath(t.TempDir()).Join("file.go")
				return path
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isAutoGenerated(tt.args(t))
			if (err != nil) != tt.wantErr {
				t.Errorf("isAutoGenerated() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("isAutoGenerated() = %v, want %v", got, tt.want)
			}
		})
	}
}
