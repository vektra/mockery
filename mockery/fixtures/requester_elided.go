package test

type RequesterElided interface {
	Get(path, url string) error
}
