package iface_typed_param

import "io"

type GetterIfaceTypedParam[T io.Reader] interface {
	Get() T
}
