package test

type VariadicFunction = func(args1 string, args2 ...interface{}) interface{}

type Variadic interface {
	VariadicFunction(str string, vFunc VariadicFunction) error
}

