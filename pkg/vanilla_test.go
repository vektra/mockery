package pkg

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"strings"
	"testing"
)

func Test_Vanilla(t *testing.T) {
	interfacePath := "vanilla_interface.go"
	interfaceName := "Vaniller"
	if !strings.Contains(interfacePath, fixturePath) {
		interfacePath = filepath.Join(fixturePath, interfacePath)
	}
	v, err := GenerateVanillaMock(interfacePath, interfaceName)

	if err != nil {
		t.Error(err)
		return
	}

	expected := `type VanillerVMock struct {
	CombinationFn func(int64) (string, error)
	IntValueFn func() int64
	StringParamFn func(string)
	VariadicFn func(abc string, more ...string) string
	WithNameFn func(abc int)
}

func (v VanillerVMock) Combination(i0 int64) (string, error) {
	return v.CombinationFn(i0)
}

func (v VanillerVMock) IntValue() (int64) {
	return v.IntValueFn()
}

func (v VanillerVMock) StringParam(s0 string) () {
	v.StringParamFn(s0)
}

func (v VanillerVMock) Variadic(abc string, more ...string) (string) {
	return v.VariadicFn(abc, more...)
}

func (v VanillerVMock) WithName(abc int) () {
	v.WithNameFn(abc)
}
`
	actual := v.Output()
	t.Logf(actual)
	assert.Equal(t, expected, actual)
}
