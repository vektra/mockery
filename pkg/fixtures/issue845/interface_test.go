package issue845

import (
	"strings"
	"testing"

	"github.com/chigopher/pathlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFix(t *testing.T) {
	for _, tt := range []struct {
		name            string
		filepath        string
		expectedPackage string
	}{
		{
			name:            "with fix",
			filepath:        "./mock_WithFix_test.go",
			expectedPackage: "package issue845_test",
		},
		{
			name:            "without fix",
			filepath:        "./mock_WithoutFix_test.go",
			expectedPackage: "package issue845",
		},
	} {
		t.Run(
			tt.name,
			func(t *testing.T) {
				path := pathlib.NewPath(tt.filepath)
				bytes, err := path.ReadFile()
				require.NoError(t, err)
				fileString := string(bytes)
				assert.True(t, strings.Contains(fileString, tt.expectedPackage))
			},
		)
	}
}
