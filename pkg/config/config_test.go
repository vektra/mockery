package config

import (
	"bytes"
	"context"
	"testing"

	"github.com/chigopher/pathlib"
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
