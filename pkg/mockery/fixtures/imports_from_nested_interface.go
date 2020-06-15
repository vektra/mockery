package test

import (
	"github.com/vektra/mockery/pkg/mockery/fixtures/http"
)

type HasConflictingNestedImports interface {
	RequesterNS
	Z() http.MyStruct
}
