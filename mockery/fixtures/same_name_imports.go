package test

import (
	"net/http"

	my_http "github.com/vektra/mockery/mockery/fixtures/http"
)

// Example is an example
type Example interface {
	A() http.Flusher
	B(my_http string) my_http.MyStruct
}
