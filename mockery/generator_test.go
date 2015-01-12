package mockery

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerator(t *testing.T) {
	parser := NewParser()
	parser.Parse(testFile)

	iface, err := parser.Find("Requester")

	gen := NewGenerator(iface)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type Requester struct {
	mock.Mock
}

func (m *Requester) Get(path string) (string, error) {
	ret := m.Called(path)

	r0 := ret.Get(0).(string)
	r1 := ret.Error(1)

	return r0, r1
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorSingleReturn(t *testing.T) {
	parser := NewParser()
	parser.Parse(testFile2)

	iface, err := parser.Find("Requester2")

	gen := NewGenerator(iface)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type Requester2 struct {
	mock.Mock
}

func (m *Requester2) Get(path string) error {
	ret := m.Called(path)

	r0 := ret.Error(0)

	return r0
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorNoArguments(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester3.go"))

	iface, err := parser.Find("Requester3")

	gen := NewGenerator(iface)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type Requester3 struct {
	mock.Mock
}

func (m *Requester3) Get() error {
	ret := m.Called()

	r0 := ret.Error(0)

	return r0
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorNoNothing(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester4.go"))

	iface, err := parser.Find("Requester4")

	gen := NewGenerator(iface)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type Requester4 struct {
	mock.Mock
}

func (m *Requester4) Get() {
	m.Called()
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorPrologue(t *testing.T) {
	parser := NewParser()
	parser.Parse(testFile)

	iface, err := parser.Find("Requester")
	assert.NoError(t, err)

	gen := NewGenerator(iface)

	gen.GeneratePrologue()

	expected := `package mocks

import "github.com/vektra/mockery/mockery/fixtures"
import "github.com/stretchr/testify/mock"

`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorProloguewithImports(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_ns.go"))

	iface, err := parser.Find("RequesterNS")
	assert.NoError(t, err)

	gen := NewGenerator(iface)

	gen.GeneratePrologue()

	expected := `package mocks

import "github.com/vektra/mockery/mockery/fixtures"
import "github.com/stretchr/testify/mock"

import "net/http"

`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorPointers(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_ptr.go"))

	iface, err := parser.Find("RequesterPtr")
	assert.NoError(t, err)

	gen := NewGenerator(iface)
	assert.NoError(t, err)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type RequesterPtr struct {
	mock.Mock
}

func (m *RequesterPtr) Get(path string) (*string, error) {
	ret := m.Called(path)

	r0 := ret.Get(0).(*string)
	r1 := ret.Error(1)

	return r0, r1
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorSlice(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_slice.go"))

	iface, err := parser.Find("RequesterSlice")
	assert.NoError(t, err)

	gen := NewGenerator(iface)
	assert.NoError(t, err)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type RequesterSlice struct {
	mock.Mock
}

func (m *RequesterSlice) Get(path string) ([]string, error) {
	ret := m.Called(path)

	r0 := ret.Get(0).([]string)
	r1 := ret.Error(1)

	return r0, r1
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorArrayLiteralLen(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_array.go"))

	iface, err := parser.Find("RequesterArray")
	assert.NoError(t, err)

	gen := NewGenerator(iface)
	assert.NoError(t, err)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type RequesterArray struct {
	mock.Mock
}

func (m *RequesterArray) Get(path string) ([2]string, error) {
	ret := m.Called(path)

	r0 := ret.Get(0).([2]string)
	r1 := ret.Error(1)

	return r0, r1
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorNamespacedTypes(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_ns.go"))

	iface, err := parser.Find("RequesterNS")
	assert.NoError(t, err)

	gen := NewGenerator(iface)
	assert.NoError(t, err)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type RequesterNS struct {
	mock.Mock
}

func (m *RequesterNS) Get(path string) (http.Response, error) {
	ret := m.Called(path)

	r0 := ret.Get(0).(http.Response)
	r1 := ret.Error(1)

	return r0, r1
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorHavingNoNamesOnArguments(t *testing.T) {
	parser := NewParser()

	parser.Parse(filepath.Join(fixturePath, "custom_error.go"))

	iface, err := parser.Find("KeyManager")
	assert.NoError(t, err)

	gen := NewGenerator(iface)
	assert.NoError(t, err)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type KeyManager struct {
	mock.Mock
}

func (m *KeyManager) GetKey(_a0 string, _a1 uint16) ([]byte, *interfaces.Err) {
	ret := m.Called(_a0, _a1)

	r0 := ret.Get(0).([]byte)
	r1 := ret.Get(1).(*interfaces.Err)

	return r0, r1
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorElidedType(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_elided.go"))

	iface, err := parser.Find("RequesterElided")

	gen := NewGenerator(iface)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type RequesterElided struct {
	mock.Mock
}

func (m *RequesterElided) Get(path string, url string) error {
	ret := m.Called(path, url)

	r0 := ret.Error(0)

	return r0
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorFuncType(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "func_type.go"))

	iface, err := parser.Find("Fooer")

	gen := NewGenerator(iface)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type Fooer struct {
	mock.Mock
}

func (m *Fooer) Foo(f func(string) string) error {
	ret := m.Called(f)

	r0 := ret.Error(0)

	return r0
}
`

	assert.Equal(t, expected, gen.buf.String())
}
