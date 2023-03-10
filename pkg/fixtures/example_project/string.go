package example_project

//go:generate --name Stringer
type Stringer interface {
    String() string
}
