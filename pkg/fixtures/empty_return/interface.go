package test

type EmptyReturn interface {
	NoArgs()
	WithArgs(a int, b string)
}
