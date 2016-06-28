package twopackages_test

import "net/http"

type InterfaceInTestPackage interface {
	A() http.File
}
