package test

type Constraint interface {
	int
}

type Generic[T Constraint] interface {
	Get() T
}

type GenericAny[T any] interface {
	Get() T
}

type GenericComparable[T comparable] interface {
	Get() T
}

type Embedded interface {
	Generic[int]
}
