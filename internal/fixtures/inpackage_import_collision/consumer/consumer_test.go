package dep

import (
	"testing"

	"github.com/stretchr/testify/assert"
	libdep "github.com/vektra/mockery/v3/internal/fixtures/inpackage_import_collision/dep"
)

func TestIfaceMock(t *testing.T) {
	want := libdep.T{}
	mockIface := NewMockIface(t)
	mockIface.EXPECT().M().Return(want)

	assert.Equal(t, want, mockIface.M())
}
