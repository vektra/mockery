package template

import "fmt"

type Packages []*Package

func (p Packages) PkgQualifier(importPath string) (string, error) {
	for _, imprt := range p {
		if imprt.Path() == "sync" {
			return imprt.Qualifier(), nil
		}
	}

	return "", fmt.Errorf("unknown import %s", importPath)
}
