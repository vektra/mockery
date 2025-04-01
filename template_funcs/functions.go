package template_funcs

import (
	"os"
	"strings"
)

func Exported(s string) string {
	if s == "" {
		return ""
	}
	for _, initialism := range golintInitialisms {
		if strings.ToUpper(s) == initialism {
			return initialism
		}
	}
	return strings.ToUpper(s[0:1]) + s[1:]
}

func ReadFile(path string) string {
	if path == "" {
		return ""
	}
	fileBytes, err := os.ReadFile(path)
	if err != nil {
		panic(err.Error())
	}
	return string(fileBytes)
}

func Add(i1 int, in ...int) int {
	var sum int = i1
	for _, i := range in {
		sum += i
	}
	return sum
}
