package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatryerParamNamesInTemplate(t *testing.T) {
	m := &MoqParamNamesInTemplate{
		SetFunc: func(mock string, callInfo int) error { return nil },
	}
	assert.NoError(t, m.Set("a", 1))

	calls := m.SetCalls()
	assert.Len(t, calls, 1)
	assert.Equal(t, "a", calls[0].Mock)
	assert.Equal(t, 1, calls[0].CallInfo)
}
