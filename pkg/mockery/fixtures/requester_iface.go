package test

import "io"

type RequesterIface interface {
	Get() io.Reader
}
