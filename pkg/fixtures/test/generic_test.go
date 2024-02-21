package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	mocks "github.com/vektra/mockery/v2/mocks/github.com/vektra/mockery/v2/pkg/fixtures"
	rtb "github.com/vektra/mockery/v2/pkg/fixtures/redefined_type_b"
)

func TestReplaceGeneric(t *testing.T) {
	type str string

	m := mocks.NewReplaceGeneric[str, str](t)

	m.EXPECT().A(rtb.B(1)).Return("")
	assert.Equal(t, m.A(rtb.B(1)), str(""))

	m.EXPECT().B().Return(2)
	assert.Equal(t, m.B(), rtb.B(2))

	m.EXPECT().C().Return("")
	assert.Equal(t, m.C(), str(""))
}

func TestReplaceGenericSelf(t *testing.T) {
	m := mocks.NewReplaceGenericSelf(t)
	m.EXPECT().A().Return(m)
	assert.Equal(t, m.A(), m)
}
