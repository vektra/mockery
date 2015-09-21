package test

type RequesterReturnElided interface {
	Get(path string) (a, b int, err error)
}
