package interfacedirrelative

import (
	"testing"

	mocks "github.com/vektra/mockery/v3/internal/fixtures/interface_dir_relative/internal/fixtures/interface_dir_relative"
)

func TestFoo(t *testing.T) {
	m := mocks.NewMockFoo(t)
	m.EXPECT().Bar().Return("foo")
	if m.Bar() != "foo" {
		t.Errorf("expected foo but got %s", m.Bar())
	}
}
