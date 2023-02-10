package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	mocks "github.com/vektra/mockery/v2/mocks/pkg/fixtures"
)

// CompatSuite covers compatibility with github.com/stretchr/testify/mock.
type CompatSuite struct {
	suite.Suite
}

// TestOnVariadicArgs asserts that methods like Mock.On accept variadic arguments
// that mirror those of the subject call.
func (s *CompatSuite) TestOnVariadicArgs() {
	t := s.T()
	m := new(mocks.RequesterVariadic)
	m.On("Sprintf", "int: %d string: %s", 22, "twenty two").Return("int: 22 string: twenty-two")
	m.Sprintf("int: %d string: %s", 22, "twenty two")
	m.AssertExpectations(t)
	m.AssertCalled(t, "Sprintf", "int: %d string: %s", 22, "twenty two")
}

// TestOnAnythingOfTypeVariadicArgs asserts that mock.AnythingOfType can be used in
// variadic arguments of methods like Mock.On.
func (s *CompatSuite) TestOnAnythingOfTypeVariadicArgs() {
	t := s.T()
	m := new(mocks.RequesterVariadic)
	m.On("Sprintf", "int: %d string: %s", mock.AnythingOfType("int"), mock.AnythingOfType("string")).Return("int: 22 string: twenty-two")
	m.Sprintf("int: %d string: %s", 22, "twenty two")
	m.AssertExpectations(s.T())
	m.AssertCalled(t, "Sprintf", "int: %d string: %s", 22, "twenty two")
}

// TestVariadicMockAnything asserts that you can use mock.Anything to match any combination of
// zero or more arguments passed to the variadic parameter of a function.
func (s *CompatSuite) TestVariadicMockAnything() {
	t := s.T()
	m := new(mocks.RequesterVariadic)
	m.On("Sprintf", mock.Anything, mock.Anything).Return("passed")
	assert.Equal(t, "passed", m.Sprintf("format string", "and", "many", "varidic", "arguments"))
	assert.Equal(t, "passed", m.Sprintf("Format string and zero variadic arguments"))
}

func TestCompatSuite(t *testing.T) {
	mockcompatSuite := new(CompatSuite)
	suite.Run(t, mockcompatSuite)
}
