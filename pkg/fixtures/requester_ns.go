package test

import "net/http"

type RequesterNS interface {
	Get(path string) (http.Response, error)
}
