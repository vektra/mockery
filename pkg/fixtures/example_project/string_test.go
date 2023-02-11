package example_project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Foo(s Stringer) string {
	return s.String()
}

func TestString(t *testing.T) {
	mockStringer := NewMockStringer(t)
	mockStringer.EXPECT().String().Return("mockery")
	assert.Equal(t, "mockery", Foo(mockStringer))
}
