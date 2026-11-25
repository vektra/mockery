package test

import (
	"errors"
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

func TestExpecterDisallowReturnAndRunAndReturn(t *testing.T) {
	const wantPanic = "Return() and RunAndReturn() cannot be used simultaneously"

	t.Run("RunAndReturn then Return panics and preserves RunAndReturn", func(t *testing.T) {
		expMock := MockExpecter{}
		var ran bool
		call := expMock.EXPECT().NoArg().RunAndReturn(func() string {
			ran = true
			return "from-run-and-return"
		})
		require.PanicsWithValue(t, wantPanic, func() {
			call.Return("from-return")
		})
		call.Once()

		require.Equal(t, "from-run-and-return", expMock.NoArg())
		require.True(t, ran)
		expMock.AssertExpectations(t)
	})

	t.Run("Return then RunAndReturn panics and preserves Return", func(t *testing.T) {
		expMock := MockExpecter{}
		var ran bool
		call := expMock.EXPECT().NoArg().Return("from-return")
		require.PanicsWithValue(t, wantPanic, func() {
			call.RunAndReturn(func() string {
				ran = true
				return "from-run-and-return"
			})
		})
		call.Once()

		require.Equal(t, "from-return", expMock.NoArg())
		require.False(t, ran)
		expMock.AssertExpectations(t)
	})

	t.Run("chained with Run in between panics in both orders", func(t *testing.T) {
		expMock := MockExpecter{}
		require.PanicsWithValue(t, wantPanic, func() {
			expMock.EXPECT().ManyArgsReturns(mock.Anything, defaultInt).
				RunAndReturn(func(string, int) ([]string, error) {
					return []string{"a"}, nil
				}).
				Run(func(string, int) {}).
				Return([]string{"b"}, nil)
		})
		require.PanicsWithValue(t, wantPanic, func() {
			expMock.EXPECT().ManyArgsReturns(mock.Anything, defaultInt).
				Return([]string{"b"}, nil).
				Run(func(string, int) {}).
				RunAndReturn(func(string, int) ([]string, error) {
					return []string{"a"}, nil
				})
		})
	})

	t.Run("NoReturn method panics in both orders", func(t *testing.T) {
		expMock := MockExpecter{}
		require.PanicsWithValue(t, wantPanic, func() {
			expMock.EXPECT().NoReturn(defaultString).
				RunAndReturn(func(string) {}).
				Return()
		})
		require.PanicsWithValue(t, wantPanic, func() {
			expMock.EXPECT().NoReturn(defaultString).
				Return().
				RunAndReturn(func(string) {})
		})
	})

	t.Run("generic mock panics in both orders", func(t *testing.T) {
		m := MockConstructorTypeParams[string, bool]{}
		require.PanicsWithValue(t, wantPanic, func() {
			m.EXPECT().Get().RunAndReturn(func() string { return "a" }).Return("b")
		})
		require.PanicsWithValue(t, wantPanic, func() {
			m.EXPECT().Get().Return("b").RunAndReturn(func() string { return "a" })
		})
	})

	t.Run("allowed combinations continue to work", func(t *testing.T) {
		expMock := MockExpecter{}

		// Repeated Return overrides previous Return.
		expMock.EXPECT().NoArg().Return("first").Return("second").Once()
		require.Equal(t, "second", expMock.NoArg())

		// Repeated RunAndReturn overrides previous RunAndReturn.
		expMock.EXPECT().NoArg().
			RunAndReturn(func() string { return "first" }).
			RunAndReturn(func() string { return "second" }).
			Once()
		require.Equal(t, "second", expMock.NoArg())

		// Return + Run on a zero-return method still works in both orders.
		var ranNoReturn int
		expMock.EXPECT().NoReturn("a").Return().Run(func(string) { ranNoReturn++ }).Once()
		expMock.EXPECT().NoReturn("b").Run(func(string) { ranNoReturn++ }).Return().Once()
		expMock.NoReturn("a")
		expMock.NoReturn("b")
		require.Equal(t, 2, ranNoReturn)

		// Run + RunAndReturn on a returning method still works in both orders.
		var sideEffects int
		expMock.EXPECT().NoArg().
			Run(func() { sideEffects++ }).
			RunAndReturn(func() string { return "rr1" }).
			Once()
		require.Equal(t, "rr1", expMock.NoArg())

		expMock.EXPECT().NoArg().
			RunAndReturn(func() string { return "rr2" }).
			Run(func() { sideEffects++ }).
			Once()
		require.Equal(t, "rr2", expMock.NoArg())
		require.Equal(t, 2, sideEffects)

		// Separate expectations on the same mock may use Return and RunAndReturn independently.
		expMock.EXPECT().ManyArgsReturns("a", 1).Return([]string{"ret"}, nil).Once()
		expMock.EXPECT().ManyArgsReturns("b", 2).RunAndReturn(func(s string, _ int) ([]string, error) {
			return []string{"rr-" + s}, nil
		}).Once()

		strs1, err := expMock.ManyArgsReturns("a", 1)
		require.NoError(t, err)
		require.Equal(t, []string{"ret"}, strs1)

		strs2, err := expMock.ManyArgsReturns("b", 2)
		require.NoError(t, err)
		require.Equal(t, []string{"rr-b"}, strs2)

		expMock.AssertExpectations(t)
	})
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
