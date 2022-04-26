package test

import "io"

type RequesterVariadic interface {
	// cases: only variadic argument, w/ and w/out interface type
	Get(values ...string) bool
	OneInterface(a ...interface{}) bool

	// cases: normal argument + variadic argument, w/ and w/o interface type
	Sprintf(format string, a ...interface{}) string
	MultiWriteToFile(filename string, w ...io.Writer) string

	// cases: returning a function with a variadic argument, w/ and w/o interface type
	LogMethodf(level string) func(message string, a ...interface{})
	LogMethodToFile(level string) func(message string, w ...io.Writer)
}
