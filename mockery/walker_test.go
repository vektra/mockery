package mockery

import (
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

type GatheringVisitor struct {
	Interfaces []*Interface
}

func (this *GatheringVisitor) VisitWalk(iface *Interface) error {
	this.Interfaces = append(this.Interfaces, iface)
	return nil
}

func NewGatheringVisitor() *GatheringVisitor {
	return &GatheringVisitor{
		Interfaces: make([]*Interface, 0, 1024),
	}
}

func TestWalkerHere(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping recursive walker test")
	}

	wd, err := os.Getwd()
	assert.NoError(t, err)
	w := Walker{
		BaseDir:   wd,
		Recursive: true,
		LimitOne:  false,
		Filter:    regexp.MustCompile(".*"),
	}

	gv := NewGatheringVisitor()

	w.Walk(gv)

	assert.True(t, len(gv.Interfaces) > 10)
	first := gv.Interfaces[0]
	assert.Equal(t, "A", first.Name)
	assert.Equal(t, getFixturePath("struct_value.go"), first.FileName)
	assert.Equal(t, "github.com/vektra/mockery/mockery/fixtures", first.QualifiedName)
}

func TestWalkerRegexp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping recursive walker test")
	}

	wd, err := os.Getwd()
	assert.NoError(t, err)
	w := Walker{
		BaseDir:   wd,
		Recursive: true,
		LimitOne:  false,
		Filter:    regexp.MustCompile(".*AsyncProducer*."),
	}

	gv := NewGatheringVisitor()

	w.Walk(gv)

	assert.True(t, len(gv.Interfaces) >= 1)
	first := gv.Interfaces[0]
	assert.Equal(t, "AsyncProducer", first.Name)
	assert.Equal(t, getFixturePath("async.go"), first.FileName)
	assert.Equal(t, "github.com/vektra/mockery/mockery/fixtures", first.QualifiedName)
}
