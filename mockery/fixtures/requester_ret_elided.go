package test

type RequesterReturnElided interface {
	Get(path string) (a, b, c int, err error)
}
