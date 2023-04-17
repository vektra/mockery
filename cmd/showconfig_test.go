package cmd

import (
	"bytes"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewShowConfigCmd(t *testing.T) {
	cmd := NewShowConfigCmd()
	assert.Equal(t, "showconfig", cmd.Name())
}

func TestShowCfg(t *testing.T) {
	v := viper.New()
	v.Set("with-expecter", true)
	cfgFile := pathlib.NewPath(t.TempDir()).Join("config.yaml")
	err := cfgFile.WriteFile([]byte(`
with-expecter: true
all: true
packages:
  github.com/vektra/mockery/v2/pkg:
    config:
      all: true`))
	assert.NoError(t, err)
	v.Set("config", cfgFile.String())
	buf := new(bytes.Buffer)
	assert.NoError(t, showConfig(nil, nil, v, buf))
	assert.Equal(t, `all: true
packages:
  github.com/vektra/mockery/v2/pkg:
    config:
      all: true
      with-expecter: true
with-expecter: true
`, buf.String())
}
