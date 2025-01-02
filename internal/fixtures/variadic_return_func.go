package test

type VariadicReturnFunc interface {
	SampleMethod(str string) func(str string, arr []int, a ...interface{})
}
