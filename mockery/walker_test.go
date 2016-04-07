package mockery

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"regexp"
	"testing"
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

	for _, i := range gv.Interfaces {
		fmt.Println(i)
	}

	assert.Equal(t, 18, len(gv.Interfaces))
	first := gv.Interfaces[0]
	assert.Equal(t, "AsyncProducer", first.Name)
	assert.Equal(t, path.Join(wd, "fixtures/async.go"), first.Path)
}

func TestWalkerRegexp(t *testing.T) {
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

	for _, i := range gv.Interfaces {
		fmt.Println(i)
	}

	assert.Equal(t, 1, len(gv.Interfaces))
	first := gv.Interfaces[0]
	assert.Equal(t, "AsyncProducer", first.Name)
	assert.Equal(t, path.Join(wd, "fixtures/async.go"), first.Path)
}
