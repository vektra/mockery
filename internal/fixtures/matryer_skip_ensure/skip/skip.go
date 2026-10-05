// Package skip is configured to skip matryer ensure declarations by default.
package skip

type First interface {
	Ping() string
}

type Second interface {
	Pong() string
}
