package template

import "fmt"

type Packages []*Package

// PkgQualifier returns the qualifier for the given pkgPath. If the pkgPath does
// not exist in the container, an error is returned.
func (p Packages) PkgQualifier(pkgPath string) (string, error) {
	for _, imprt := range p {
		if imprt.Path() == pkgPath {
			return imprt.Qualifier(), nil
		}
	}

	return "", fmt.Errorf("unknown import %s", pkgPath)
}
