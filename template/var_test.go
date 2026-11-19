package template

import (
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNillableUnsafePointer(t *testing.T) {
	assert.True(t, Var{typ: types.Typ[types.UnsafePointer]}.Nillable())
	assert.False(t, Var{typ: types.Typ[types.Uintptr]}.Nillable())
}
