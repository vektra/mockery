package test

type ConsulLock interface {
	Lock(<-chan struct{}) (<-chan struct{}, error)
	Unlock() error
}
