package example_project

//go:generate mockery --name GoGenerateExample --all=False
type GoGenerateExample interface {
	Foo(s string) error
}
