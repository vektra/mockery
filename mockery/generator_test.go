package mockery

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func run(g *Generator, name string) error {
	err := g.Setup(name)
	if err != nil {
		return err
	}

	return g.Generate()
}

func TestGenerator(t *testing.T) {
	parser := NewParser()
	parser.Parse(testFile)

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	err := run(gen, "Requester")
	assert.NoError(t, err)

	expected := `type Requester struct {
	mock.Mock
}

func (m *Requester) Get(path string) (string, error) {
	ret := m.Called(path)

	r0 := m.Get(0).(string)
	r1 := m.Get(1).(error)

	return r0, r1
}
`

	assert.Equal(t, expected, out.String())
}

func TestGeneratorSingleReturn(t *testing.T) {
	parser := NewParser()
	parser.Parse(testFile2)

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	err := run(gen, "Requester2")
	assert.NoError(t, err)

	expected := `type Requester2 struct {
	mock.Mock
}

func (m *Requester2) Get(path string) error {
	ret := m.Called(path)

	r0 := m.Get(0).(error)

	return r0
}
`

	assert.Equal(t, expected, out.String())
}

func TestGeneratorNoArguments(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester3.go"))

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	err := run(gen, "Requester3")
	assert.NoError(t, err)

	expected := `type Requester3 struct {
	mock.Mock
}

func (m *Requester3) Get() error {
	ret := m.Called()

	r0 := m.Get(0).(error)

	return r0
}
`

	assert.Equal(t, expected, out.String())
}

func TestGeneratorNoNothing(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester4.go"))

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	err := run(gen, "Requester4")
	assert.NoError(t, err)

	expected := `type Requester4 struct {
	mock.Mock
}

func (m *Requester4) Get() {
	m.Called()
}
`

	assert.Equal(t, expected, out.String())
}

func TestGeneratorPrologue(t *testing.T) {
	parser := NewParser()
	parser.Parse(testFile)

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	gen.GeneratePrologue()

	expected := `package mocks

import "github.com/stretchr/testify/mock"

`

	assert.Equal(t, expected, out.String())
}

func TestGeneratorProloguewithImports(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_ns.go"))

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	gen.GeneratePrologue()

	expected := `package mocks

import "github.com/stretchr/testify/mock"

import "net/http"

`

	assert.Equal(t, expected, out.String())
}

func TestGeneratorPointers(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_ptr.go"))

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	err := run(gen, "RequesterPtr")
	assert.NoError(t, err)

	expected := `type RequesterPtr struct {
	mock.Mock
}

func (m *RequesterPtr) Get(path string) (*string, error) {
	ret := m.Called(path)

	r0 := m.Get(0).(*string)
	r1 := m.Get(1).(error)

	return r0, r1
}
`

	assert.Equal(t, expected, out.String())
}

func TestGeneratorSlice(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_slice.go"))

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	err := run(gen, "RequesterSlice")
	assert.NoError(t, err)

	expected := `type RequesterSlice struct {
	mock.Mock
}

func (m *RequesterSlice) Get(path string) ([]string, error) {
	ret := m.Called(path)

	r0 := m.Get(0).([]string)
	r1 := m.Get(1).(error)

	return r0, r1
}
`

	assert.Equal(t, expected, out.String())
}

func TestGeneratorArrayLiteralLen(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_array.go"))

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	err := run(gen, "RequesterArray")
	assert.NoError(t, err)

	expected := `type RequesterArray struct {
	mock.Mock
}

func (m *RequesterArray) Get(path string) ([2]string, error) {
	ret := m.Called(path)

	r0 := m.Get(0).([2]string)
	r1 := m.Get(1).(error)

	return r0, r1
}
`

	assert.Equal(t, expected, out.String())
}

func TestGeneratorNamespacedTypes(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_ns.go"))

	var out bytes.Buffer

	gen := NewGenerator(parser, &out)

	err := run(gen, "RequesterNS")
	assert.NoError(t, err)

	expected := `type RequesterNS struct {
	mock.Mock
}

func (m *RequesterNS) Get(path string) (http.Response, error) {
	ret := m.Called(path)

	r0 := m.Get(0).(http.Response)
	r1 := m.Get(1).(error)

	return r0, r1
}
`

	assert.Equal(t, expected, out.String())
}
