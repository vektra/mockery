package templateexercise

import (
	"context"

	"golang.org/x/exp/constraints"
)

// GenDecl comments
type (
	// Exercise is an interface that is used to render a template that exercises
	// all parts of the template data passed to the template.
	Exercise[T any, Ordered constraints.Ordered] interface {
		// Foo is a foo
		Foo(ctx context.Context, typeParam T, ord1, ord2 Ordered) error
	} // This is a line comment
)
