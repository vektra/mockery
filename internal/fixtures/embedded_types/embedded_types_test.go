package embeddedtypes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmbeddedTypeInPackage(t *testing.T) {
	var server GreeterServer = &MoqGreeterServerLocal{
		GreetFunc: func(_ context.Context, name string) (string, error) {
			return name, nil
		},
	}
	got, err := server.Greet(t.Context(), "Alice")
	require.NoError(t, err)
	require.Equal(t, "Alice", got)
}
