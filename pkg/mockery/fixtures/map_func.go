package test

type MapFunc interface {
	Get(m map[string]func(string) string) error
}
