package stackerr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStackErr(t *testing.T) {
	err := assert.AnError

	s, ok := GetStack(err)
	assert.False(t, ok)
	assert.Empty(t, s)

	err = NewStackErr(err)
	assert.Equal(t, assert.AnError.Error(), err.Error())

	s, ok = GetStack(err)
	assert.True(t, ok)
	assert.NotEmpty(t, s)

	err = NewStackErr(fmt.Errorf("wrapped error can still get stack: %w", err))
	s, ok = GetStack(err)
	assert.True(t, ok)
	assert.NotEmpty(t, s)
}

func TestStackErrf(t *testing.T) {
	err := assert.AnError

	s, ok := GetStack(err)
	assert.False(t, ok)
	assert.Empty(t, s)

	err = NewStackErrf(err, "error message %d %s", 1, "a")
	assert.Equal(t, "error message 1 a: "+assert.AnError.Error(), err.Error())

	s, ok = GetStack(err)
	assert.True(t, ok)
	assert.NotEmpty(t, s)
}
