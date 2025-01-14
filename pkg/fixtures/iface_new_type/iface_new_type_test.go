package iface_new_type_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vektra/mockery/v2/pkg"
)

func TestParsing(t *testing.T) {
	parser := pkg.NewParser(nil)
	ctx := context.Background()
	require.NoError(t, parser.ParsePackages(ctx, []string{"github.com/vektra/mockery/v2/pkg/fixtures/iface_new_type"}))
	require.NoError(t, parser.Load(ctx))

	for _, ifaceName := range []string{"Interface1", "Interface2", "Interface3", "Interface4"} {
		iface, err := parser.Find(ifaceName)
		require.NoError(t, err)
		require.NotNil(t, iface)
	}
}

func TestUsage(t *testing.T) {
	interface1 := NewMockInterface1(t)
	interface1.EXPECT().Method1().Return()
	interface1.Method1()

	interface2 := NewMockInterface2(t)
	interface2.EXPECT().Method1().Return()
	interface2.Method1()

	interface3 := NewMockInterface3(t)
	interface3.EXPECT().Method1().Return()
	interface3.Method1()

	interface4 := NewMockInterface4(t)
	interface4.EXPECT().Method1().Return()
	interface4.Method1()
}
