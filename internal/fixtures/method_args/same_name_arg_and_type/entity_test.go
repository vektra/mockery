package same_name_arg_and_type

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testStruct struct {
	interfaceA interfaceA
}

func (s *testStruct) ExecDoB() interfaceB {
	var in interfaceB = nil
	return s.interfaceA.DoB(in)
}

func (s *testStruct) ExecDoB0() interfaceB0 {
	var in interfaceB0 = nil
	return s.interfaceA.DoB0(in)
}

func (s *testStruct) ExecDoB0v2() interfaceB0 {
	var in interfaceB0 = nil
	return s.interfaceA.DoB0v2(in)
}

func Test(t *testing.T) {
	t.Run("ExecDoB", func(t *testing.T) {
		mockInterfaceB := newmockinterfaceB(t)
		mockInterfaceA := newmockinterfaceA(t)
		mockInterfaceA.On("DoB", mock.Anything).Return(mockInterfaceB)

		s := testStruct{
			interfaceA: mockInterfaceA,
		}
		res := s.ExecDoB()
		assert.Equal(t, mockInterfaceB, res)
	})
	t.Run("ExecDoB0", func(t *testing.T) {
		mockInterfaceB0 := newmockinterfaceB0(t)
		mockInterfaceA := newmockinterfaceA(t)
		mockInterfaceA.On("DoB0", mock.Anything).Return(mockInterfaceB0)

		s := testStruct{
			interfaceA: mockInterfaceA,
		}
		res := s.ExecDoB0()
		assert.Equal(t, mockInterfaceB0, res)
	})
	t.Run("ExecDoB0v2", func(t *testing.T) {
		mockInterfaceB0 := newmockinterfaceB0(t)
		mockInterfaceA := newmockinterfaceA(t)
		mockInterfaceA.On("DoB0v2", mock.Anything).Return(mockInterfaceB0)

		s := testStruct{
			interfaceA: mockInterfaceA,
		}
		res := s.ExecDoB0v2()
		assert.Equal(t, mockInterfaceB0, res)
	})
}
