package test

type Requester interface {
	Get(path string) (string, error)
}
