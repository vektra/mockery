package config

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			configFile := pathlib.NewPath(t.TempDir()).Join("config.yaml")
			require.NoError(t, configFile.WriteFile([]byte(tt.config)))

			flags := pflag.NewFlagSet("test", pflag.ExitOnError)
			flags.String("config", "", "")

			require.NoError(t, flags.Parse([]string{"--config", configFile.String()}))

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
