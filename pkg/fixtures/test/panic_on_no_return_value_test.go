package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mocks "github.com/vektra/mockery/v2/mocks/github.com/vektra/mockery/v2/pkg/fixtures"
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
		if r := recover(); r != nil {
			require.NotNil(t, r)
			assert.Equal(t, "no return value specified for DoSomething", r.(string))
		}
	}()

	m.DoSomething()

}
