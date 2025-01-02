package test

type RequesterArray interface {
	Get(path string) ([2]string, error)
}
