package pkg

import (
	"go/types"
)

type Method struct {
	Name      string
	Signature *types.Signature
}
