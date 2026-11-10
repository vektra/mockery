package iface_typed_param_lowercase

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIfaceWithIfaceTypedParamLowerCaseReturnValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		arg       *int
		returnVal *int
	}{
		{"nil return val", nil, nil},
		{"returning val", toPtr(2), toPtr(2)},
	}
	for _, test := range tests {
		t.Run(test.name, func(st *testing.T) {
			m := NewMockGetterIfaceTypedParam[*int](st)
			m.EXPECT().Get(test.arg).Return(test.returnVal)

			assert.Equal(st, test.returnVal, m.Get(test.arg))
		})
	}
}

func toPtr(i int) *int {
	return &i
}

func TestMatryerIfaceWithIfaceTypedParamLowerCase(t *testing.T) {
	m := &MoqGetterIfaceTypedParam[string]{
		GetFunc: func(v string) string { return v + "!" },
	}

	assert.Equal(t, "a!", m.Get("a"))
	assert.Len(t, m.GetCalls(), 1)
}
