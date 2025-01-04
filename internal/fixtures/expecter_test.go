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
	expMock := NewMockExpecter(t)

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
		}, "call did not panic")
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
