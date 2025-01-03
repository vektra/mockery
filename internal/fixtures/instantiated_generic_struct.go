package test

// Tests that mockery does not try to generate mocks for a struct type.
type InstantiatedStruct GenericStruct[int]

type GenericStruct[T any] struct {
	Attribute []T
}
