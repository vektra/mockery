package templateexercise

import (
	"context"

	"golang.org/x/exp/constraints"
)

type Exercise[T any, Ordered constraints.Ordered] interface {
	Foo(ctx context.Context, typeParam T, ordered Ordered) error
}
