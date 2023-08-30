package test

type Expecter interface {
	NoArg() string
	NoReturn(str string)
	ManyArgsReturns(str string, i int) (strs []string, err error)
	Variadic(ints ...int) error
	VariadicMany(i int, a string, intfs ...interface{}) error
	VariadicNoReturn(j int, is ...interface{})
}
