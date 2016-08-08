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

func NewErr(msg string, code uint64) Err {
	return Err{msg, code}
}

type KeyManager interface {
	GetKey(string, uint16) ([]byte, *Err)
}
