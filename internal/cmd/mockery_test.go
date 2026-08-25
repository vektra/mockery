package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartCPUProfile(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "cpu.prof")

	stop, err := startCPUProfile(profilePath)
	require.NoError(t, err)
	require.NoError(t, stop())

	info, err := os.Stat(profilePath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0))
}

func TestStartCPUProfileBadPath(t *testing.T) {
	profilePath := filepath.Join(t.TempDir(), "does-not-exist", "cpu.prof")

	_, err := startCPUProfile(profilePath)
	assert.Error(t, err)
}
