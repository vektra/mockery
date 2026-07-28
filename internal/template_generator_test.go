package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindPkgPath(t *testing.T) {
	pkgPath, err := findPkgPath("./fixtures")
	require.NoError(t, err)
	assert.NotEmpty(t, pkgPath)
}
