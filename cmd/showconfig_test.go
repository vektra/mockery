package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewShowConfigCmd(t *testing.T) {
	cmd := NewShowConfigCmd()
	assert.Equal(t, "showconfig", cmd.Name())
}
