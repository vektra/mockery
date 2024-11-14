package pkg

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	mocks "github.com/vektra/mockery/v2/mocks/github.com/vektra/mockery/v2/pkg/fixtures"
	test "github.com/vektra/mockery/v2/pkg/fixtures"
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

func (s *CompatSuite) TestOnEmptyReturn() {
	m := mocks.NewEmptyReturn(s.T())
	var target test.EmptyReturn = m

	s.Run("NoArgs", func() {
		run := false

		m.EXPECT().NoArgs().RunAndReturn(func() {
			run = true
		})

		target.NoArgs()

		s.True(run)
	})

	s.Run("WithArgs", func() {
		run := false

		m.EXPECT().WithArgs(42, "foo").RunAndReturn(func(arg0 int, arg1 string) {
			run = true
			s.Equal(42, arg0)
			s.Equal("foo", arg1)
		})

		target.WithArgs(42, "foo")

		s.True(run)
	})
}

func TestCompatSuite(t *testing.T) {
	mockcompatSuite := new(CompatSuite)
	suite.Run(t, mockcompatSuite)
}
