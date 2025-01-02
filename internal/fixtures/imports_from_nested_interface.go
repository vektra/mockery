package test

import (
	"github.com/vektra/mockery/v3/internal/fixtures/http"
)

type HasConflictingNestedImports interface {
	RequesterNS
	Z() http.MyStruct
}
