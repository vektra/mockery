// Code generated by mockery; DO NOT EDIT.
// github.com/vektra/mockery

package mocks

import (
	"io"

	mock "github.com/stretchr/testify/mock"
	test "github.com/vektra/mockery/v2/pkg/fixtures"
	"github.com/vektra/mockery/v2/pkg/fixtures/constraints"
)

// NewRequesterGenerics creates a new instance of RequesterGenerics. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRequesterGenerics[TAnyParam any, TComparableParam comparable, TSignedParam constraints.Signed, TIntfParam test.GetInt, TExternalIntfParam io.Writer, TGenIntfParam test.GetGeneric[TSigned], TInlineTypeParam interface{ ~int | ~uint }, TInlineTypeGenericParam interface {
	~int | test.GenericType[int, test.GetInt]
	comparable
}](t interface {
	mock.TestingT
	Cleanup(func())
}) *RequesterGenerics[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	mock := &RequesterGenerics[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

// RequesterGenerics is an autogenerated mock type for the RequesterGenerics type
type RequesterGenerics[TAnyParam any, TComparableParam comparable, TSignedParam constraints.Signed, TIntfParam test.GetInt, TExternalIntfParam io.Writer, TGenIntfParam test.GetGeneric[TSigned], TInlineTypeParam interface{ ~int | ~uint }, TInlineTypeGenericParam interface {
	~int | test.GenericType[int, test.GetInt]
	comparable
}] struct {
	mock.Mock
}

type RequesterGenerics_Expecter[TAnyParam any, TComparableParam comparable, TSignedParam constraints.Signed, TIntfParam test.GetInt, TExternalIntfParam io.Writer, TGenIntfParam test.GetGeneric[TSigned], TInlineTypeParam interface{ ~int | ~uint }, TInlineTypeGenericParam interface {
	~int | test.GenericType[int, test.GetInt]
	comparable
}] struct {
	mock *mock.Mock
}

func (_m *RequesterGenerics[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) EXPECT() *RequesterGenerics_Expecter[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	return &RequesterGenerics_Expecter[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]{mock: &_m.Mock}
}

// GenericAnonymousStructs provides a mock function for the type RequesterGenerics
func (_mock *RequesterGenerics[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) GenericAnonymousStructs(valParam struct{ Type1 TExternalIntf }) struct {
	Type2 test.GenericType[string, test.EmbeddedGet[int]]
} {
	ret := _mock.Called(valParam)

	if len(ret) == 0 {
		panic("no return value specified for GenericAnonymousStructs")
	}

	var r0 struct {
		Type2 test.GenericType[string, test.EmbeddedGet[int]]
	}
	if returnFunc, ok := ret.Get(0).(func(struct{ Type1 TExternalIntf }) struct {
		Type2 test.GenericType[string, test.EmbeddedGet[int]]
	}); ok {
		r0 = returnFunc(valParam)
	} else {
		r0 = ret.Get(0).(struct {
			Type2 test.GenericType[string, test.EmbeddedGet[int]]
		})
	}
	return r0
}

// RequesterGenerics_GenericAnonymousStructs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GenericAnonymousStructs'
type RequesterGenerics_GenericAnonymousStructs_Call[TAnyParam any, TComparableParam comparable, TSignedParam constraints.Signed, TIntfParam test.GetInt, TExternalIntfParam io.Writer, TGenIntfParam test.GetGeneric[TSigned], TInlineTypeParam interface{ ~int | ~uint }, TInlineTypeGenericParam interface {
	~int | test.GenericType[int, test.GetInt]
	comparable
}] struct {
	*mock.Call
}

// GenericAnonymousStructs is a helper method to define mock.On call
//   - valParam
func (_e *RequesterGenerics_Expecter[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) GenericAnonymousStructs(valParam interface{}) *RequesterGenerics_GenericAnonymousStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	return &RequesterGenerics_GenericAnonymousStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]{Call: _e.mock.On("GenericAnonymousStructs", valParam)}
}

func (_c *RequesterGenerics_GenericAnonymousStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) Run(run func(valParam struct{ Type1 TExternalIntf })) *RequesterGenerics_GenericAnonymousStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(struct{ Type1 TExternalIntf }))
	})
	return _c
}

func (_c *RequesterGenerics_GenericAnonymousStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) Return(valOutParam struct {
	Type2 test.GenericType[string, test.EmbeddedGet[int]]
}) *RequesterGenerics_GenericAnonymousStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	_c.Call.Return(valOutParam)
	return _c
}

func (_c *RequesterGenerics_GenericAnonymousStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) RunAndReturn(run func(valParam struct{ Type1 TExternalIntf }) struct {
	Type2 test.GenericType[string, test.EmbeddedGet[int]]
}) *RequesterGenerics_GenericAnonymousStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	_c.Call.Return(run)
	return _c
}

// GenericArguments provides a mock function for the type RequesterGenerics
func (_mock *RequesterGenerics[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) GenericArguments(vParam TAny, v_1_Param TComparable) (TSigned, TIntf) {
	ret := _mock.Called(vParam, v_1_Param)

	if len(ret) == 0 {
		panic("no return value specified for GenericArguments")
	}

	var r0 TSigned
	var r1 TIntf
	if returnFunc, ok := ret.Get(0).(func(TAny, TComparable) (TSigned, TIntf)); ok {
		return returnFunc(vParam, v_1_Param)
	}
	if returnFunc, ok := ret.Get(0).(func(TAny, TComparable) TSigned); ok {
		r0 = returnFunc(vParam, v_1_Param)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(TSigned)
		}
	}
	if returnFunc, ok := ret.Get(1).(func(TAny, TComparable) TIntf); ok {
		r1 = returnFunc(vParam, v_1_Param)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(TIntf)
		}
	}
	return r0, r1
}

// RequesterGenerics_GenericArguments_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GenericArguments'
type RequesterGenerics_GenericArguments_Call[TAnyParam any, TComparableParam comparable, TSignedParam constraints.Signed, TIntfParam test.GetInt, TExternalIntfParam io.Writer, TGenIntfParam test.GetGeneric[TSigned], TInlineTypeParam interface{ ~int | ~uint }, TInlineTypeGenericParam interface {
	~int | test.GenericType[int, test.GetInt]
	comparable
}] struct {
	*mock.Call
}

// GenericArguments is a helper method to define mock.On call
//   - vParam
//   - v_1_Param
func (_e *RequesterGenerics_Expecter[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) GenericArguments(vParam interface{}, v_1_Param interface{}) *RequesterGenerics_GenericArguments_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	return &RequesterGenerics_GenericArguments_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]{Call: _e.mock.On("GenericArguments", vParam, v_1_Param)}
}

func (_c *RequesterGenerics_GenericArguments_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) Run(run func(vParam TAny, v_1_Param TComparable)) *RequesterGenerics_GenericArguments_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(TAny), args[1].(TComparable))
	})
	return _c
}

func (_c *RequesterGenerics_GenericArguments_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) Return(vOutParam TSigned, vOut_1_Param TIntf) *RequesterGenerics_GenericArguments_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	_c.Call.Return(vOutParam, vOut_1_Param)
	return _c
}

func (_c *RequesterGenerics_GenericArguments_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) RunAndReturn(run func(vParam TAny, v_1_Param TComparable) (TSigned, TIntf)) *RequesterGenerics_GenericArguments_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	_c.Call.Return(run)
	return _c
}

// GenericStructs provides a mock function for the type RequesterGenerics
func (_mock *RequesterGenerics[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) GenericStructs(genericTypeParam test.GenericType[TAny, TIntf]) test.GenericType[TSigned, TIntf] {
	ret := _mock.Called(genericTypeParam)

	if len(ret) == 0 {
		panic("no return value specified for GenericStructs")
	}

	var r0 test.GenericType[TSigned, TIntf]
	if returnFunc, ok := ret.Get(0).(func(test.GenericType[TAny, TIntf]) test.GenericType[TSigned, TIntf]); ok {
		r0 = returnFunc(genericTypeParam)
	} else {
		r0 = ret.Get(0).(test.GenericType[TSigned, TIntf])
	}
	return r0
}

// RequesterGenerics_GenericStructs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GenericStructs'
type RequesterGenerics_GenericStructs_Call[TAnyParam any, TComparableParam comparable, TSignedParam constraints.Signed, TIntfParam test.GetInt, TExternalIntfParam io.Writer, TGenIntfParam test.GetGeneric[TSigned], TInlineTypeParam interface{ ~int | ~uint }, TInlineTypeGenericParam interface {
	~int | test.GenericType[int, test.GetInt]
	comparable
}] struct {
	*mock.Call
}

// GenericStructs is a helper method to define mock.On call
//   - genericTypeParam
func (_e *RequesterGenerics_Expecter[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) GenericStructs(genericTypeParam interface{}) *RequesterGenerics_GenericStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	return &RequesterGenerics_GenericStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]{Call: _e.mock.On("GenericStructs", genericTypeParam)}
}

func (_c *RequesterGenerics_GenericStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) Run(run func(genericTypeParam test.GenericType[TAny, TIntf])) *RequesterGenerics_GenericStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(test.GenericType[TAny, TIntf]))
	})
	return _c
}

func (_c *RequesterGenerics_GenericStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) Return(genericTypeOutParam test.GenericType[TSigned, TIntf]) *RequesterGenerics_GenericStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	_c.Call.Return(genericTypeOutParam)
	return _c
}

func (_c *RequesterGenerics_GenericStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam]) RunAndReturn(run func(genericTypeParam test.GenericType[TAny, TIntf]) test.GenericType[TSigned, TIntf]) *RequesterGenerics_GenericStructs_Call[TAnyParam, TComparableParam, TSignedParam, TIntfParam, TExternalIntfParam, TGenIntfParam, TInlineTypeParam, TInlineTypeGenericParam] {
	_c.Call.Return(run)
	return _c
}
