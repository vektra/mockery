package pkg

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	mocks "github.com/vektra/mockery/v2/mocks/pkg/fixtures"
)

// CompatSuite covers compatbility with github.com/stretchr/testify/mock.
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

func TestCompatSuite(t *testing.T) {
	mockcompatSuite := new(CompatSuite)
	suite.Run(t, mockcompatSuite)
}
