package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTestifyReturnFuncParamNames(t *testing.T) {
	m := NewMockReturnFuncParamNames(t)
	m.EXPECT().Set("a", false).RunAndReturn(func(key string, ok bool) (int, error) {
		assert.False(t, ok)
		return 1, nil
	})
	m.EXPECT().Enabled(false).RunAndReturn(func(ok bool) bool {
		return ok
	})
	m.EXPECT().Apply(mock.Anything).RunAndReturn(func(returnFunc func() int) int {
		return returnFunc()
	})

	n, err := m.Set("a", false)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.False(t, m.Enabled(false))
	assert.Equal(t, 7, m.Apply(func() int { return 7 }))
}
