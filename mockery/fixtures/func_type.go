package test

type Fooer interface {
	Foo(f func(x string) string) error
	Bar(f func([]int))
}
