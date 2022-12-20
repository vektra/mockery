package test

type RequesterSlice interface {
	Get(path string) ([]string, error)
}
