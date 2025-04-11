package unexported

import (
	"strings"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnexportedConstructorName(t *testing.T) {
	mockFile := pathlib.NewPath("./mocks_testify_unexported_test.go")
	b, err := mockFile.ReadFile()
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(b), "func newmockfoo("))
}
