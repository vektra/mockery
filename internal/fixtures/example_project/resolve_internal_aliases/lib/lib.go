// Package lib is the public face of a library whose real definitions live in lib/internal.
package lib

import "github.com/vektra/mockery/v3/internal/fixtures/example_project/resolve_internal_aliases/lib/internal"

type (
	Options = internal.Options
	Result  = internal.Result
	Service = internal.Service
)
