package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFoo(t *testing.T) {
	m := newMockfoo(t)
	m.EXPECT().Bar().Return("foo")
	assert.Equal(t, baz("foo"), m.Bar())
}
