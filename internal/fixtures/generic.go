package test

import (
	"io"

	"github.com/vektra/mockery/v3/internal/fixtures/constraints"
)

type RequesterGenerics[
	TAny any,
	TComparable comparable,
	TSigned constraints.Signed, // external constraint
	TIntf GetInt, // internal interface
	TExternalIntf io.Writer, // external interface
	TGenIntf GetGeneric[TSigned], // generic interface
	TInlineType interface{ ~int | ~uint }, // inlined interface constraints
	TInlineTypeGeneric interface {
		~int | GenericType[int, GetInt]
		comparable
	}, // inlined type constraints
] interface {
	GenericArguments(TAny, TComparable) (TSigned, TIntf)
	GenericStructs(GenericType[TAny, TIntf]) GenericType[TSigned, TIntf]
	GenericAnonymousStructs(struct{ Type1 TExternalIntf }) struct {
		Type2 GenericType[string, EmbeddedGet[int]]
	}
}

type GenericType[T any, S GetInt] struct {
	Any  T
	Some []S
}

type GetInt interface{ Get() int }

type GetGeneric[T constraints.Integer] interface{ Get() T }

type EmbeddedGet[T constraints.Signed] interface{ GetGeneric[T] }

type ReplaceGeneric[
	TImport any,
	TConstraint constraints.Signed,
	TKeep any,
] interface {
	A(t1 TImport) TKeep
	B() TImport
	C() TConstraint
}

type ReplaceGenericSelf[T any] interface {
	A() T
}
