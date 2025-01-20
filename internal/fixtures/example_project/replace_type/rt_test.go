package replace_type

import (
	"strings"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceType(t *testing.T) {
	mockFile := pathlib.NewPath("./mocks_replace_type_test.go")
	b, err := mockFile.ReadFile()
	require.NoError(t, err)
	// .mockery.yml replaced github.com/vektra/mockery/v3/internal/fixtures/example_project/replace_type/rti/rt1
	// with github.com/vektra/mockery/v3/internal/fixtures/example_project/replace_type/rti/rt2
	assert.True(t, strings.Contains(string(b), "*RTypeReplaced1) Replace1(f rt2.RType2) {"))
	// This should contain no replaced type.
	assert.True(t, strings.Contains(string(b), "*MockRType) Replace1(f rt1.RType1) {"))
}
