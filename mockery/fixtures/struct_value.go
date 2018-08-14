package test

type B struct{}

type A interface {
	Call() (B, error)
}
