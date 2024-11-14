package iface_typed_param_test

import (
	"bufio"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIfaceWithIfaceTypedParamReturnValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		returnVal *bufio.Reader
	}{
		{"nil return val", nil},
		{"returning val", bufio.NewReader(http.NoBody)},
	}
	for _, test := range tests {
		t.Run(test.name, func(st *testing.T) {
			m := NewGetterIfaceTypedParam[*bufio.Reader](st)
			m.EXPECT().Get().Return(test.returnVal)

			assert.Equal(st, test.returnVal, m.Get())
		})
	}
}
