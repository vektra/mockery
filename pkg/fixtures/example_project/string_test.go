package example_project_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vektra/mockery/v2/pkg/fixtures/example_project"
)

func Foo(s example_project.Stringer) string {
	return s.String()
}

func TestString(t *testing.T) {
	mockStringer := NewMockStringer(t)
	mockStringer.EXPECT().String().Return("mockery")
	assert.Equal(t, "mockery", Foo(mockStringer))
}
