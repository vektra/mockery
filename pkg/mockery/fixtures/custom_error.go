package test

type Err struct {
	msg  string
	code uint64
}

func (e *Err) Error() string {
	return e.msg
}

func (e *Err) Code() uint64 {
	return e.code
}

type KeyManager interface {
	GetKey(string, uint16) ([]byte, *Err)
}
