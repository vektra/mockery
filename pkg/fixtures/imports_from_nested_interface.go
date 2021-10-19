package test

import (
	"github.com/pendo-io/b2h-mockgen/pkg/fixtures/http"
)

type HasConflictingNestedImports interface {
	RequesterNS
	Z() http.MyStruct
}
