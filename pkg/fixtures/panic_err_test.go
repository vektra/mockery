package test_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocks "github.com/vektra/mockery/v3/mocks/github.com/vektra/mockery/v3/pkg/fixtures"
)

func TestPanicOnNoReturnValue(t *testing.T) {
	m := mocks.NewPanicOnNoReturnValue(t)
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
