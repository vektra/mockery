package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	mocks "github.com/vektra/mockery/v3/mocks/github.com/vektra/mockery/v3/internal/fixtures"
)

func TestVariadicReturnFunc(t *testing.T) {
	m := mocks.NewVariadicReturnFunc(t)
	m.EXPECT().SampleMethod("").Return(func(s string, l []int, a ...any) {
		assert.Equal(t, "foo", s)
		assert.Equal(t, []int{1, 2, 3}, l)
		assert.Equal(t, []any{"one", "two", "three"}, a)
	})
	m.SampleMethod("")("foo", []int{1, 2, 3}, "one", "two", "three")
}
