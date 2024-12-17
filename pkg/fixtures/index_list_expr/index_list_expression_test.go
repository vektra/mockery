package index_list_expr_test

import (
	"context"
	"testing"

	"github.com/vektra/mockery/v2/pkg"

	"github.com/stretchr/testify/require"
)

func TestParsing(t *testing.T) {
	parser := pkg.NewParser(nil)
	ctx := context.Background()
	require.NoError(t, parser.ParsePackages(ctx, []string{"github.com/vektra/mockery/v2/pkg/fixtures/index_list_expr"}))
	require.NoError(t, parser.Load(ctx))

	for _, ifaceName := range []string{"GenericMultipleTypes", "IndexListExpr"} {
		iface, err := parser.Find(ifaceName)
		require.NoError(t, err)
		require.NotNil(t, iface)
	}
}

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
