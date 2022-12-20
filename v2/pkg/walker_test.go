package pkg

import (
	"bytes"
	"context"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

type GatheringVisitor struct {
	Interfaces []*Interface
}

func (v *GatheringVisitor) VisitWalk(ctx context.Context, iface *Interface) error {
	v.Interfaces = append(v.Interfaces, iface)
	return nil
}

func NewGatheringVisitor() *GatheringVisitor {
	return &GatheringVisitor{
		Interfaces: make([]*Interface, 0, 1024),
	}
}

type BufferedProvider struct {
	buf *bytes.Buffer
}

func NewBufferedProvider() *BufferedProvider {
	return &BufferedProvider{
		buf: new(bytes.Buffer),
	}
}

func (bp *BufferedProvider) String() string {
	return bp.buf.String()
}

func (bp *BufferedProvider) GetWriter(context.Context, *Interface) (io.Writer, error, Cleanup) {
	return bp.buf, nil, func() error { return nil }
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

	w.Walk(context.Background(), gv)

	assert.True(t, len(gv.Interfaces) > 10)
	first := gv.Interfaces[0]
	assert.Equal(t, "A", first.Name)
	assert.Equal(t, getFixturePath("struct_value.go"), first.FileName)
	assert.Equal(t, "github.com/vektra/mockery/v2/pkg/fixtures", first.QualifiedName)
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

	w.Walk(context.Background(), gv)

	assert.True(t, len(gv.Interfaces) >= 1)
	first := gv.Interfaces[0]
	assert.Equal(t, "AsyncProducer", first.Name)
	assert.Equal(t, getFixturePath("async.go"), first.FileName)
	assert.Equal(t, "github.com/vektra/mockery/v2/pkg/fixtures", first.QualifiedName)
}

func TestPackagePrefix(t *testing.T) {
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

	bufferedProvider := NewBufferedProvider()
	gv := &GeneratorVisitor{
		InPackage:         false,
		Osp:               bufferedProvider,
		PackageName:       "mocks",
		PackageNamePrefix: "prefix_test_",
	}

	w.Walk(context.Background(), gv)
	assert.Regexp(t, regexp.MustCompile("package prefix_test_test"), bufferedProvider.String())
}
