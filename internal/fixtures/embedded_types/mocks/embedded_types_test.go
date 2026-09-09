package embeddedmocks

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektra/mockery/v3/internal/fixtures/embedded_types"
	"github.com/vektra/mockery/v3/internal/fixtures/embedded_types/metadata"
)

func TestEmbeddedTypes(t *testing.T) {
	mock := &MoqGreeterServer{
		GreetFunc: func(_ context.Context, name string) (string, error) {
			return "Hello, " + name, nil
		},
		Metadata: metadata.Metadata{Name: "test"},
	}
	var server embeddedtypes.GreeterServer = mock
	got, err := server.Greet(t.Context(), "Alice")
	require.NoError(t, err)
	require.Equal(t, "Hello, Alice", got)
	require.Len(t, mock.GreetCalls(), 1)
	require.Empty(t, mock.mustEmbedUnimplementedGreeterServerCalls())
	require.Equal(t, "test", mock.Metadata.Name)
	field, ok := reflect.TypeFor[MoqGreeterServer]().FieldByName("Label")
	require.True(t, ok)
	require.Equal(t, "{{label}}", field.Tag.Get("json"))

	var other embeddedtypes.OtherServer = &MoqOtherServer{VersionFunc: func() string { return "v1" }}
	require.Equal(t, "v1", other.Version())

	var store embeddedtypes.Store[string] = &MoqStore[string]{LoadFunc: func() string { return "value" }}
	require.Equal(t, "value", store.Load())
}
