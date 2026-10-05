// Package ensure is configured to emit matryer ensure declarations by default.
package ensure

type First interface {
	Ping() string
}

type Value struct {
	N int
}

type Typed interface {
	Get() Value
}
