package index_list_expr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUsage(t *testing.T) {
	gmt := NewMockGenericMultipleTypes[string, int, bool](t)
	testString := "foo"
	gmt.EXPECT().Func(&testString, 1).Return(false)
	require.Equal(t, false, gmt.Func(&testString, 1))

	ile := NewMockIndexListExpr(t)
	testInt := 1
	ile.EXPECT().Func(&testInt, "foo").Return(true)
	require.Equal(t, true, ile.Func(&testInt, "foo"))
}
