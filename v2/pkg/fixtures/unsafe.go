package test

import "unsafe"

type UnsafeInterface interface {
	Do(ptr *unsafe.Pointer)
}
