package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedConfig string = `all: false
dir: '{{.InterfaceDir}}'
filename: mocks_test.go
force-file-write: true
formatter: goimports
formatter-options:
  goimports:
    all-errors: false
    comments: true
    format-only: true
    fragment: false
    local-prefix: ""
    tab-indent: true
    tab-width: 8
generate: true
include-auto-generated: false
log-level: info
structname: '{{.Mock}}{{.InterfaceName}}'
pkgname: '{{.SrcPackageName}}'
recursive: false
require-template-schema-exists: true
template: testify
template-schema: '{{.Template}}.schema.json'
packages:
  github.com/org/repo:
    config:
      all: true
`

func Test_initRun(t *testing.T) {
	type args struct {
		args   []string
		params func(t *testing.T, configPath string) argGetter
	}
	tests := []struct {
		name       string
		configPath string
		args       args
	}{
		{
			name: "specify --config case",
			args: args{
				args: []string{"github.com/org/repo"},
				params: func(t *testing.T, configPath string) argGetter {
					m := newMockargGetter(t)
					m.EXPECT().GetString("config").Return(configPath, nil)
					return m
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			config := filepath.Join(tmpDir, "out.yml")
			initRun(tt.args.args, tt.args.params(t, config))

			b, err := os.ReadFile(config)
			require.NoError(t, err)
			assert.Equal(t, expectedConfig, string(b))
		})
	}
}
