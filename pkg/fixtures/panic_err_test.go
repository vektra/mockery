package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPanicOnNoReturnValue(t *testing.T) {
	m := NewMockPanicOnNoReturnValue(t)
	m.EXPECT().DoSomething()

	var panicOccurred bool
	defer func() {
		assert.True(t, panicOccurred)
	}()
	defer func() {
		panicOccurred = true

		r := recover()
		require.NotNil(t, r)
		assert.Equal(t, "no return value specified for DoSomething", r.(string))
	}()

	m.DoSomething()
}
