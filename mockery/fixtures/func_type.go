package test

type Fooer interface {
	Foo(f func(x string) string) error
	Bar(f func([]int))
	Baz(path string) func(x string) string
}
