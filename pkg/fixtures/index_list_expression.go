package test

type GenericMultipleTypes[T1 any, T2 any, T3 any] interface {
	Func(arg1 *T1, arg2 T2) T3
}

type IndexListExpr GenericMultipleTypes[int, string, bool]
