package empty_return

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	m := NewMockEmptyReturn(t)
	var target EmptyReturn = m

	t.Run("NoArgs", func(t *testing.T) {
		run := false

		m.EXPECT().NoArgs().RunAndReturn(func() {
			run = true
		})

		target.NoArgs()

		require.True(t, run)
	})

	t.Run("WithArgs", func(t *testing.T) {
		run := false

		m.EXPECT().WithArgs(42, "foo").RunAndReturn(func(arg0 int, arg1 string) {
			run = true
			require.Equal(t, 42, arg0)
			require.Equal(t, "foo", arg1)
		})

		target.WithArgs(42, "foo")

		require.True(t, run)
	})
}

func TestMatryerNoReturnStub(t *testing.T) {
	m := &StubMatyerEmptyReturn{}
	// If this is a stub, this should not panic.
	m.NoArgs()
}
