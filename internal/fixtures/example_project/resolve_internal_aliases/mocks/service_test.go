package mocks

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vektra/mockery/v3/internal/fixtures/example_project/resolve_internal_aliases/lib"
)

// The mock is generated outside the lib tree, so Go would reject any import of lib/internal.
// This file compiling at all is the main check; the rest confirms the mock is usable as lib.Service.
var _ lib.Service = (*MockService)(nil)

func TestMockServiceUsesPublicAliases(t *testing.T) {
	svc := NewMockService(t)
	svc.EXPECT().
		Do(mock.Anything, lib.Options{Name: "a"}).
		Return(lib.Result{Value: 1}, nil)
	svc.EXPECT().
		Many(mock.Anything, mock.Anything).
		Return(map[string]*lib.Options{"k": {Name: "b"}})

	res, err := svc.Do(context.Background(), lib.Options{Name: "a"})
	require.NoError(t, err)
	require.Equal(t, 1, res.Value)

	got := svc.Many([]lib.Result{{Value: 2}}, func(lib.Options) {})
	require.Equal(t, "b", got["k"].Name)
}

func TestGeneratedMockDoesNotImportInternal(t *testing.T) {
	b, err := os.ReadFile("./mocks_testify_resolve_internal_aliases_test.go")
	require.NoError(t, err)
	require.False(t, strings.Contains(string(b), "lib/internal"))
}
