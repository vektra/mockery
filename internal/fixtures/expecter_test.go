package test

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	defaultString = "some input string"
	defaultInt    = 1
	defaultError  = errors.New("some error")
)

// Test that the generated code for ExpecterTest interface is usable
func TestExpecter(t *testing.T) {
	expMock := MockExpecter{}

	t.Run("NoArg", func(t *testing.T) {
		var runCalled bool
		expMock.EXPECT().NoArg().Run(func() {
			runCalled = true
		}).Return(defaultString).Once()

		// Good call
		str := expMock.NoArg()
		require.Equal(t, defaultString, str)
		require.True(t, runCalled)

		// Call again panic
		assert.Panics(t, func() {
			expMock.NoArg()
		})
		expMock.AssertExpectations(t)
	})

	t.Run("NoReturn", func(t *testing.T) {
		var runCalled bool
		expMock.EXPECT().NoReturn(mock.Anything).Run(func(s string) {
			require.Equal(t, defaultString, s)
			runCalled = true
		}).Return().Once()

		// Good call
		expMock.NoReturn(defaultString)
		require.True(t, runCalled)

		// Call again panic
		require.Panics(t, func() {
			expMock.NoReturn(defaultString)
		})
		expMock.AssertExpectations(t)
	})

	t.Run("ManyArgsReturns", func(t *testing.T) {
		var runCalled bool
		expMock.EXPECT().ManyArgsReturns(mock.Anything, defaultInt).Run(func(s string, i int) {
			require.Equal(t, defaultString, s)
			require.Equal(t, defaultInt, i)
			runCalled = true
		}).Return([]string{defaultString, defaultString}, defaultError).Once()

		// Call with wrong arg
		require.Panics(t, func() {
			_, _ = expMock.ManyArgsReturns(defaultString, 0)
		})

		// Good call
		strs, err := expMock.ManyArgsReturns(defaultString, defaultInt)
		require.Equal(t, []string{defaultString, defaultString}, strs)
		require.Equal(t, defaultError, err)
		require.True(t, runCalled)

		// Call again panic
		require.Panics(t, func() {
			_, _ = expMock.ManyArgsReturns(defaultString, defaultInt)
		})
		expMock.AssertExpectations(t)
	})

	t.Run("Variadic", func(t *testing.T) {
		runCalled := 0

		expMock.EXPECT().Variadic(1).Run(func(ints ...int) {
			require.Equal(t, []int{1}, ints)
			runCalled++
		}).Return(defaultError).Once()

		expMock.EXPECT().Variadic(1, 2, 3).Run(func(ints ...int) {
			require.Equal(t, []int{1, 2, 3}, ints)
			runCalled++
		}).Return(nil).Once()

		expMock.EXPECT().Variadic(1, mock.Anything, 3, mock.Anything).Run(func(ints ...int) {
			require.Equal(t, []int{1, 2, 3, 4}, ints)
			runCalled++
		}).Return(nil).Once()

		expMock.EXPECT().Variadic([]interface{}{2, 3, mock.Anything}...).Run(func(ints ...int) {
			require.Equal(t, []int{2, 3, 4}, ints)
			runCalled++
		}).Return(nil).Once()

		args := []int{1, 2, 3, 4, 5}
		expMock.EXPECT().Variadic(intfSlice(args)...).Run(func(ints ...int) {
			require.Equal(t, args, ints)
			runCalled++
		}).Return(nil).Once()

		require.Error(t, expMock.Variadic(1))
		require.NoError(t, expMock.Variadic(1, 2, 3))
		require.NoError(t, expMock.Variadic(1, 2, 3, 4))
		require.NoError(t, expMock.Variadic(2, 3, 4))
		require.NoError(t, expMock.Variadic(args...))
		require.Equal(t, 5, runCalled)
		expMock.AssertExpectations(t)
	})

	t.Run("VariadicOtherArgs", func(t *testing.T) {
		runCalled := 0

		expMock.EXPECT().VariadicMany(defaultInt, defaultString).Return(defaultError).
			Run(func(i int, a string, intfs ...interface{}) {
				require.Equal(t, defaultInt, i)
				require.Equal(t, defaultString, a)
				require.Empty(t, intfs)
				runCalled++
			}).Once()
		require.Error(t, expMock.VariadicMany(defaultInt, defaultString))

		expMock.EXPECT().VariadicMany(defaultInt, defaultString, 1).Return(defaultError).
			Run(func(i int, a string, intfs ...interface{}) {
				require.Equal(t, defaultInt, i)
				require.Equal(t, defaultString, a)
				require.Equal(t, []interface{}{1}, intfs)
				runCalled++
			}).Once()
		require.Error(t, expMock.VariadicMany(defaultInt, defaultString, 1))

		expMock.EXPECT().VariadicMany(mock.Anything, mock.Anything, 1, nil, mock.AnythingOfType("string")).Return(nil).
			Run(func(i int, a string, intfs ...interface{}) {
				require.Equal(t, defaultInt, i)
				require.Equal(t, defaultString, a)
				require.Equal(t, []interface{}{1, nil, "blah"}, intfs)
				runCalled++
			}).Once()
		require.Panics(t, func() {
			assert.NoError(t, expMock.VariadicMany(defaultInt, defaultString, 1, nil, 123))
		})
		require.NoError(t, expMock.VariadicMany(defaultInt, defaultString, 1, nil, "blah"))

		expMock.EXPECT().VariadicMany(mock.Anything, mock.Anything, 1, nil, "blah").Run(func(i int, a string, intfs ...interface{}) {
			require.Equal(t, defaultInt, i)
			require.Equal(t, defaultString, a)
			require.Equal(t, []interface{}{1, nil, "blah"}, intfs)
			runCalled++
		}).Return(defaultError).Once()
		require.Panics(t, func() {
			assert.NoError(t, expMock.VariadicMany(defaultInt, defaultString, 1, nil, "other string"))
		})
		err := expMock.VariadicMany(defaultInt, defaultString, 1, nil, "blah")
		require.Equal(t, defaultError, err)

		args := []interface{}{1, 2, 3, 4, 5}
		expMock.EXPECT().VariadicMany(defaultInt, defaultString, args...).Run(func(i int, a string, intfs ...interface{}) {
			require.Equal(t, defaultInt, i)
			require.Equal(t, defaultString, a)
			require.Equal(t, []interface{}{1, 2, 3, 4, 5}, intfs)
			runCalled++
		}).Return(nil).Once()
		require.NoError(t, expMock.VariadicMany(defaultInt, defaultString, args...))

		require.Equal(t, 5, runCalled)
		expMock.AssertExpectations(t)
	})
}

func TestExpecterAssertNotCalled(t *testing.T) {
	t.Run("NoArg", func(t *testing.T) {
		expMock := MockExpecter{}
		passT := &assertNotCalledTestingT{}
		assert.True(t, expMock.ASSERT(passT).NotCalled().NoArg())
		assert.Empty(t, passT.errors)
		assert.Equal(t, 2, passT.helperCalls)

		expMock.EXPECT().NoArg().Return(defaultString).Once()
		require.Equal(t, defaultString, expMock.NoArg())

		fakeT := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeT).NotCalled().NoArg())
		require.Len(t, fakeT.errors, 1)
		assert.Contains(t, fakeT.errors[0], `Expected "NoArg" to not have been called`)
		assert.GreaterOrEqual(t, fakeT.helperCalls, 2)

		expMock.AssertExpectations(t)
	})

	t.Run("NoReturn", func(t *testing.T) {
		expMock := MockExpecter{}
		assert.True(t, expMock.ASSERT(t).NotCalled().NoReturn(mock.Anything))

		expMock.EXPECT().NoReturn(defaultString).Return().Once()
		expMock.NoReturn(defaultString)

		assert.True(t, expMock.ASSERT(t).NotCalled().NoReturn("other string"))

		fakeT := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeT).NotCalled().NoReturn(defaultString))
		require.Len(t, fakeT.errors, 1)
		assert.Contains(t, fakeT.errors[0], `Expected "NoReturn" to not have been called`)

		fakeTAny := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeTAny).NotCalled().NoReturn(mock.Anything))
		require.Len(t, fakeTAny.errors, 1)

		expMock.AssertExpectations(t)
	})

	t.Run("ManyArgsReturns", func(t *testing.T) {
		expMock := MockExpecter{}
		expMock.EXPECT().ManyArgsReturns(defaultString, defaultInt).Return([]string{defaultString}, nil).Once()

		strs, err := expMock.ManyArgsReturns(defaultString, defaultInt)
		require.NoError(t, err)
		require.Equal(t, []string{defaultString}, strs)

		assert.True(t, expMock.ASSERT(t).NotCalled().ManyArgsReturns(defaultString, 999))
		assert.True(t, expMock.ASSERT(t).NotCalled().ManyArgsReturns("other", defaultInt))

		fakeExact := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeExact).NotCalled().ManyArgsReturns(defaultString, defaultInt))
		require.Len(t, fakeExact.errors, 1)

		fakeAny := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeAny).NotCalled().ManyArgsReturns(mock.Anything, defaultInt))
		require.Len(t, fakeAny.errors, 1)

		fakeMatched := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeMatched).NotCalled().ManyArgsReturns(
			mock.AnythingOfType("string"),
			mock.MatchedBy(func(i int) bool { return i == defaultInt }),
		))
		require.Len(t, fakeMatched.errors, 1)

		expMock.AssertExpectations(t)
	})

	t.Run("VariadicUnrolled", func(t *testing.T) {
		expMock := MockExpecter{}
		assert.True(t, expMock.ASSERT(t).NotCalled().Variadic())
		assert.True(t, expMock.ASSERT(t).NotCalled().Variadic(1, 2, 3))

		expMock.EXPECT().Variadic(1, 2, 3).Return(nil).Once()
		require.NoError(t, expMock.Variadic(1, 2, 3))

		assert.True(t, expMock.ASSERT(t).NotCalled().Variadic())
		assert.True(t, expMock.ASSERT(t).NotCalled().Variadic(1, 2))
		assert.True(t, expMock.ASSERT(t).NotCalled().Variadic(1, 2, 4))

		fakeExact := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeExact).NotCalled().Variadic(1, 2, 3))
		require.Len(t, fakeExact.errors, 1)

		fakeAny := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeAny).NotCalled().Variadic(1, mock.Anything, 3))
		require.Len(t, fakeAny.errors, 1)

		expMock.EXPECT().Variadic().Return(nil).Once()
		require.NoError(t, expMock.Variadic())

		fakeEmpty := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeEmpty).NotCalled().Variadic())
		require.Len(t, fakeEmpty.errors, 1)

		expMock.AssertExpectations(t)
	})

	t.Run("VariadicManyUnrolled", func(t *testing.T) {
		expMock := MockExpecter{}
		expMock.EXPECT().VariadicMany(defaultInt, defaultString, 1, nil, "x").Return(nil).Once()
		require.NoError(t, expMock.VariadicMany(defaultInt, defaultString, 1, nil, "x"))

		notCalled := expMock.ASSERT(t).NotCalled()
		assert.True(t, notCalled.VariadicMany(defaultInt, defaultString))
		assert.True(t, notCalled.VariadicMany(defaultInt, defaultString, 1, nil, "y"))

		fakeExact := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeExact).NotCalled().VariadicMany(defaultInt, defaultString, 1, nil, "x"))
		require.Len(t, fakeExact.errors, 1)

		fakeMatched := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeMatched).NotCalled().VariadicMany(
			mock.Anything,
			defaultString,
			1,
			nil,
			mock.AnythingOfType("string"),
		))
		require.Len(t, fakeMatched.errors, 1)

		expMock.EXPECT().VariadicMany(defaultInt, defaultString).Return(nil).Once()
		require.NoError(t, expMock.VariadicMany(defaultInt, defaultString))

		fakeNoVar := &assertNotCalledTestingT{}
		assert.False(t, expMock.ASSERT(fakeNoVar).NotCalled().VariadicMany(defaultInt, defaultString))
		require.Len(t, fakeNoVar.errors, 1)

		expMock.AssertExpectations(t)
	})

	t.Run("VariadicRolled", func(t *testing.T) {
		rolledMock := MockExpecterAndRolledVariadic{}
		rolledMock.EXPECT().Variadic([]int{1, 2}).Return(nil).Once()
		require.NoError(t, rolledMock.Variadic(1, 2))

		assert.True(t, rolledMock.ASSERT(t).NotCalled().Variadic([]int{1, 3}))

		fakeExact := &assertNotCalledTestingT{}
		assert.False(t, rolledMock.ASSERT(fakeExact).NotCalled().Variadic([]int{1, 2}))
		require.Len(t, fakeExact.errors, 1)

		fakeAny := &assertNotCalledTestingT{}
		assert.False(t, rolledMock.ASSERT(fakeAny).NotCalled().Variadic(mock.Anything))
		require.Len(t, fakeAny.errors, 1)

		rolledMock.EXPECT().VariadicMany(defaultInt, defaultString, []any{1, "x"}).Return(nil).Once()
		require.NoError(t, rolledMock.VariadicMany(defaultInt, defaultString, 1, "x"))

		assert.True(t, rolledMock.ASSERT(t).NotCalled().VariadicMany(defaultInt, defaultString, []any{1, "y"}))

		fakeMany := &assertNotCalledTestingT{}
		assert.False(t, rolledMock.ASSERT(fakeMany).NotCalled().VariadicMany(defaultInt, defaultString, []any{1, "x"}))
		require.Len(t, fakeMany.errors, 1)

		rolledMock.AssertExpectations(t)
	})

	t.Run("GenericMocks", func(t *testing.T) {
		typeParamMock := NewMockConstructorTypeParams[string, bool](t)
		assert.True(t, typeParamMock.ASSERT(t).NotCalled().Get())

		typeParamMock.EXPECT().Get().Return("value").Once()
		require.Equal(t, "value", typeParamMock.Get())

		fakeT := &assertNotCalledTestingT{}
		assert.False(t, typeParamMock.ASSERT(fakeT).NotCalled().Get())
		require.Len(t, fakeT.errors, 1)

		genericMock := NewMockReplaceGeneric[string, int, bool](t)
		genericMock.EXPECT().A("called").Return(true).Once()
		require.True(t, genericMock.A("called"))

		assert.True(t, genericMock.ASSERT(t).NotCalled().A("other"))
		assert.True(t, genericMock.ASSERT(t).NotCalled().B())

		fakeGeneric := &assertNotCalledTestingT{}
		assert.False(t, genericMock.ASSERT(fakeGeneric).NotCalled().A("called"))
		require.Len(t, fakeGeneric.errors, 1)
	})

	t.Run("ParamNamesCollisionSafety", func(t *testing.T) {
		m := NewMockParamNamesInTemplate(t)
		m.EXPECT().Set("a", 1).Return(nil).Once()
		require.NoError(t, m.Set("a", 1))

		assert.True(t, m.ASSERT(t).NotCalled().Set("b", 1))

		fakeT := &assertNotCalledTestingT{}
		assert.False(t, m.ASSERT(fakeT).NotCalled().Set("a", 1))
		require.Len(t, fakeT.errors, 1)
	})

	t.Run("TestingTWithoutHelper", func(t *testing.T) {
		expMock := MockExpecter{}
		legacyT := &constructorTestingT{}
		assert.True(t, expMock.ASSERT(legacyT).NotCalled().NoArg())
		assert.Empty(t, legacyT.errors)

		expMock.EXPECT().NoArg().Return(defaultString).Once()
		require.Equal(t, defaultString, expMock.NoArg())

		assert.False(t, expMock.ASSERT(legacyT).NotCalled().NoArg())
		require.Len(t, legacyT.errors, 1)

		expMock.AssertExpectations(t)
	})

	t.Run("BackwardCompatibilityWithTestifyAssertNotCalled", func(t *testing.T) {
		expMock := MockExpecter{}
		expMock.EXPECT().ManyArgsReturns(defaultString, defaultInt).Return([]string{defaultString}, nil).Once()
		_, err := expMock.ManyArgsReturns(defaultString, defaultInt)
		require.NoError(t, err)

		assert.True(t, expMock.AssertNotCalled(t, "NoArg"))
		assert.True(t, expMock.AssertNotCalled(t, "ManyArgsReturns", "other", 99))
		assert.True(t, expMock.ASSERT(t).NotCalled().NoArg())
		assert.True(t, expMock.ASSERT(t).NotCalled().ManyArgsReturns("other", 99))

		fakeDirect := &assertNotCalledTestingT{}
		assert.False(t, expMock.AssertNotCalled(fakeDirect, "ManyArgsReturns", defaultString, defaultInt))
		require.Len(t, fakeDirect.errors, 1)

		expMock.AssertExpectations(t)
	})
}

type assertNotCalledTestingT struct {
	helperCalls int
	errors      []string
}

func (t *assertNotCalledTestingT) Helper() {
	t.helperCalls++
}

func (t *assertNotCalledTestingT) Logf(string, ...any) {}

func (t *assertNotCalledTestingT) Errorf(format string, args ...any) {
	t.errors = append(t.errors, fmt.Sprintf(format, args...))
}

func (t *assertNotCalledTestingT) FailNow() {
	panic("unexpected FailNow")
}

func intfSlice(slice interface{}) []interface{} {
	val := reflect.ValueOf(slice)
	switch val.Kind() {
	case reflect.Slice, reflect.Array, reflect.String:
		out := make([]interface{}, val.Len())
		for i := 0; i < val.Len(); i++ {
			out[i] = val.Index(i).Interface()
		}
		return out
	default:
		panic("inftSlice only accepts slices or arrays")
	}
}
