package foo

type Baz struct {
	One string
	Two int
}

type Foo interface {
	DoFoo() string
	GetBaz() (*Baz, error)
}
