package template

import "go/types"

type TypeParam struct {
	Param
	Constraint types.Type
}
