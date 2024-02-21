package inpackage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMoq(t *testing.T) {
	mock := FooMoq{
		GetFunc: func(key ArgType) ReturnType {
			return ReturnType(key + "suffix")
		},
	}
	assert.Equal(t, mock.Get("foo"), ReturnType("foosuffix"))
}
