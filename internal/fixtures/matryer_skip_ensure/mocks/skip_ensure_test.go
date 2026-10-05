package skipensuremocks

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/mockery/v3/internal/fixtures/matryer_skip_ensure/ensure"
	"github.com/vektra/mockery/v3/internal/fixtures/matryer_skip_ensure/skip"
)

func TestInterfaceSkipEnsureMocks(t *testing.T) {
	var interfaceSkip ensure.First = &MoqFirstSkipEnsure{PingFunc: func() string { return "skip" }}
	require.Equal(t, "skip", interfaceSkip.Ping())
	require.Len(t, interfaceSkip.(*MoqFirstSkipEnsure).PingCalls(), 1)

	var interfaceEnsure skip.First = &MoqFirstEnsure{PingFunc: func() string { return "ensure" }}
	require.Equal(t, "ensure", interfaceEnsure.Ping())
	require.Len(t, interfaceEnsure.(*MoqFirstEnsure).PingCalls(), 1)

	var mixedFirst skip.First = &MoqFirstMixed{PingFunc: func() string { return "ping" }}
	var mixedSecond skip.Second = &MoqSecondMixed{PongFunc: func() string { return "pong" }}
	require.Equal(t, "ping", mixedFirst.Ping())
	require.Equal(t, "pong", mixedSecond.Pong())
	require.Len(t, mixedFirst.(*MoqFirstMixed).PingCalls(), 1)
	require.Len(t, mixedSecond.(*MoqSecondMixed).PongCalls(), 1)

	var typed ensure.Typed = &MoqTyped{GetFunc: func() ensure.Value { return ensure.Value{N: 7} }}
	require.Equal(t, ensure.Value{N: 7}, typed.Get())
	require.Len(t, typed.(*MoqTyped).GetCalls(), 1)
}

func TestInterfaceSkipEnsureDeclarations(t *testing.T) {
	const (
		ensureImport = `"github.com/vektra/mockery/v3/internal/fixtures/matryer_skip_ensure/ensure"`
		skipImport   = `"github.com/vektra/mockery/v3/internal/fixtures/matryer_skip_ensure/skip"`
	)

	tests := []struct {
		name     string
		filepath string
		present  []string
		absent   []string
	}{
		{
			name:     "interface disables ensure",
			filepath: "./mocks_matryer_interface_skip_test.go",
			absent:   []string{ensureImport, "var _ ensure.First"},
		},
		{
			name:     "interface enables ensure",
			filepath: "./mocks_matryer_interface_ensure_test.go",
			present:  []string{skipImport, "var _ skip.First = &MoqFirstEnsure{}"},
		},
		{
			name:     "mixed shared output",
			filepath: "./mocks_matryer_mixed_test.go",
			present:  []string{skipImport, "var _ skip.First = &MoqFirstMixed{}"},
			absent:   []string{"var _ skip.Second"},
		},
		{
			name:     "method retains source import",
			filepath: "./mocks_matryer_method_import_test.go",
			present:  []string{ensureImport},
			absent:   []string{"var _ ensure.Typed"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := os.ReadFile(tt.filepath)
			require.NoError(t, err)
			for _, s := range tt.present {
				assert.Contains(t, string(b), s)
			}
			for _, s := range tt.absent {
				assert.NotContains(t, string(b), s)
			}
		})
	}
}
