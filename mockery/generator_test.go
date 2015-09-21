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

func (_m *Requester) Get(path string) (string, error) {
	ret := _m.Called(path)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

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

func (_m *Requester2) Get(path string) error {
	ret := _m.Called(path)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Error(0)
	}

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

func (_m *Requester3) Get() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

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

func (_m *Requester4) Get() {
	_m.Called()
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorUnexported(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_unexported.go"))

	iface, err := parser.Find("requester")

	gen := NewGenerator(iface)
	gen.ip = true

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type mockRequester struct {
	mock.Mock
}

func (_m *mockRequester) Get() {
	_m.Called()
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

	gen.GeneratePrologue("mocks")

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

	gen.GeneratePrologue("mocks")

	expected := `package mocks

import "github.com/vektra/mockery/mockery/fixtures"
import "github.com/stretchr/testify/mock"

import "net/http"

`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorPrologueNote(t *testing.T) {
	parser := NewParser()
	parser.Parse(testFile)

	iface, err := parser.Find("Requester")
	assert.NoError(t, err)

	gen := NewGenerator(iface)

	gen.GeneratePrologueNote("A\\nB")

	expected := `
// A
// B

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

func (_m *RequesterPtr) Get(path string) (*string, error) {
	ret := _m.Called(path)

	var r0 *string
	if rf, ok := ret.Get(0).(func(string) *string); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

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

func (_m *RequesterSlice) Get(path string) ([]string, error) {
	ret := _m.Called(path)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

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

func (_m *RequesterArray) Get(path string) ([2]string, error) {
	ret := _m.Called(path)

	var r0 [2]string
	if rf, ok := ret.Get(0).(func(string) [2]string); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([2]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

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

func (_m *RequesterNS) Get(path string) (http.Response, error) {
	ret := _m.Called(path)

	var r0 http.Response
	if rf, ok := ret.Get(0).(func(string) http.Response); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(http.Response)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Error(1)
	}

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

func (_m *KeyManager) GetKey(_a0 string, _a1 uint16) ([]byte, *test.Err) {
	ret := _m.Called(_a0, _a1)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(string, uint16) []byte); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 *test.Err
	if rf, ok := ret.Get(1).(func(string, uint16) *test.Err); ok {
		r1 = rf(_a0, _a1)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*test.Err)
		}
	}

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

func (_m *RequesterElided) Get(path string, url string) error {
	ret := _m.Called(path, url)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(path, url)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorReturnElidedType(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_ret_elided.go"))

	iface, err := parser.Find("RequesterReturnElided")

	gen := NewGenerator(iface)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type RequesterReturnElided struct {
	mock.Mock
}

func (_m *RequesterReturnElided) Get(path string) (int, int, int, error) {
	ret := _m.Called(path)

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(path)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 int
	if rf, ok := ret.Get(1).(func(string) int); ok {
		r1 = rf(path)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 int
	if rf, ok := ret.Get(2).(func(string) int); ok {
		r2 = rf(path)
	} else {
		r2 = ret.Get(2).(int)
	}

	var r3 error
	if rf, ok := ret.Get(3).(func(string) error); ok {
		r3 = rf(path)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorVariableArgs(t *testing.T) {

	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "requester_variable.go"))

	iface, err := parser.Find("RequesterVariable")

	gen := NewGenerator(iface)

	err = gen.Generate()
	assert.NoError(t, err)
	expected := `type RequesterVariable struct {
	mock.Mock
}

func (_m *RequesterVariable) Get(values ...string) bool {
	ret := _m.Called(values)

	var r0 bool
	if rf, ok := ret.Get(0).(func(...string) bool); ok {
		r0 = rf(values...)
	} else {
		r0 = ret.Get(0).(bool)
	}

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

func (_m *Fooer) Foo(f func(string) string) error {
	ret := _m.Called(f)

	var r0 error
	if rf, ok := ret.Get(0).(func(func(string) string) error); ok {
		r0 = rf(f)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
func (_m *Fooer) Bar(f func([]int) ) {
	_m.Called(f)
}
func (_m *Fooer) Baz(path string) func(string) string {
	ret := _m.Called(path)

	var r0 func(string) string
	if rf, ok := ret.Get(0).(func(string) func(string) string); ok {
		r0 = rf(path)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(func(string) string)
		}
	}

	return r0
}
`

	assert.Equal(t, expected, gen.buf.String())
}

func TestGeneratorChanType(t *testing.T) {
	parser := NewParser()
	parser.Parse(filepath.Join(fixturePath, "async.go"))

	iface, err := parser.Find("AsyncProducer")

	gen := NewGenerator(iface)

	err = gen.Generate()
	assert.NoError(t, err)

	expected := `type AsyncProducer struct {
	mock.Mock
}

func (_m *AsyncProducer) Input() chan<- bool {
	ret := _m.Called()

	var r0 chan<- bool
	if rf, ok := ret.Get(0).(func() chan<- bool); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan<- bool)
		}
	}

	return r0
}
func (_m *AsyncProducer) Output() <-chan bool {
	ret := _m.Called()

	var r0 <-chan bool
	if rf, ok := ret.Get(0).(func() <-chan bool); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan bool)
		}
	}

	return r0
}
func (_m *AsyncProducer) Whatever() chan bool {
	ret := _m.Called()

	var r0 chan bool
	if rf, ok := ret.Get(0).(func() chan bool); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan bool)
		}
	}

	return r0
}
`

	assert.Equal(t, expected, gen.buf.String())
}
