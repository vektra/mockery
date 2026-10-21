package internal

import "context"

type Options struct{ Name string }

type Result struct{ Value int }

// Service is declared here and re-exported by package lib.
type Service interface {
	Do(ctx context.Context, opts Options) (Result, error)
	Many(results []Result, each func(Options)) map[string]*Options
}
