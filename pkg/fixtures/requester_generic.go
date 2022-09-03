package test

type RequesterGeneric[T any] interface {
	Get(path T) (T, error)
}
