package empty_return

type EmptyReturn interface {
	NoArgs()
	WithArgs(a int, b string)
}
