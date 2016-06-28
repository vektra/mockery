package twopackagestest

import "net/http"

type InterfaceInTestPackage interface {
	A() http.File
}
