package test

type RequesterPtr interface {
	Get(path string) (*string, error)
}
