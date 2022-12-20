package example_project

import "github.com/vektra/mockery/v3/pkg/fixtures/example_project/foo"

type Root interface {
	TakesBaz(*foo.Baz)
	ReturnsFoo() (foo.Foo, error)
}
