package test

import (
	"github.com/vektra/mockery/v3/pkg/fixtures/http"
)

type HasConflictingNestedImports interface {
	RequesterNS
	Z() http.MyStruct
}
