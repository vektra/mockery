package template

import "go/types"

type TypeParamData struct {
	ParamData
	Constraint types.Type
}
