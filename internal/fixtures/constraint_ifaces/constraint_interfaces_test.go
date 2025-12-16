package skipconstraintifaces

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektra/mockery/v3/internal/file"
)

func TestSkipConstraintInterfaces(t *testing.T) {
	exists, err := file.Exists("./mocks_testify_skipconstraintifaces_test")

	require.NoError(t, err)
	require.False(t, exists)
}
