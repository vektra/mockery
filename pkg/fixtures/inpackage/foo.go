package inpackage

type ArgType string

type ReturnType string

type Foo interface {
	Get(key ArgType) ReturnType
}
