package test

import (
	"github.com/vektra/mockery/v2/pkg/fixtures/http"
)

type HasConflictingNestedImports interface {
	RequesterNS
	Z() http.MyStruct
}
