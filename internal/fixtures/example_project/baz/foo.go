package baz

import (
	ifoo "github.com/vektra/mockery/v3/internal/fixtures/example_project/baz/internal/foo"
)

type Baz = ifoo.InternalBaz

type Foo interface {
	DoFoo() string
	GetBaz() (*Baz, error)
}
