package test

type GenericInterface[M any] interface {
	Func(arg *M) int
}

type InstantiatedGenericInterface GenericInterface[float32]
