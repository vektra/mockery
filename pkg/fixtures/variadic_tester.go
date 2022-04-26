package test

type Level string

type VariadicTester interface {
	LogMethod(level Level) func(string)
	LogMethodf(level Level) func(message string, a ...interface{})
}
