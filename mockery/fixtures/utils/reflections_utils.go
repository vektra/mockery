package utils

import (
	"reflect"
	"runtime"
	"strings"
)

func GetFunctionName(function interface{}) string {
	absoluteFunctionName := runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
	slices := strings.Split(absoluteFunctionName, ".")
	slices = strings.Split(slices[len(slices)-1], "-")

	return slices[0]
}
