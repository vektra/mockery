package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()
	assert.Equal(t, "mockery", cmd.Name())
}
