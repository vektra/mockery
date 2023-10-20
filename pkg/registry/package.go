package registry

import (
	"path"
	"strings"
)

type TypesPackage interface {
	Name() string
	Path() string
}

// Package represents an imported package.
type Package struct {
	pkg TypesPackage

	Alias string
}

// NewPackage creates a new instance of Package.
func NewPackage(pkg TypesPackage) *Package {
	return &Package{pkg: pkg}
}

// Qualifier returns the qualifier which must be used to refer to types
// declared in the package.
func (p *Package) Qualifier() string {
	if p == nil {
		return ""
	}

	if p.Alias != "" {
		return p.Alias
	}

	return p.pkg.Name()
}

// Path is the full package import path (without vendor).
func (p *Package) Path() string {
	if p == nil {
		return ""
	}

	return stripVendorPath(p.pkg.Path())
}

var replacer = strings.NewReplacer(
	"go-", "",
	"-go", "",
	"-", "",
	"_", "",
	".", "",
	"@", "",
	"+", "",
	"~", "",
)

// uniqueName generates a unique name for a package by concatenating
// path components. The generated name is guaranteed to unique with an
// appropriate level because the full package import paths themselves
// are unique.
func (p Package) uniqueName(lvl int) string {
	pp := strings.Split(p.Path(), "/")
	reverse(pp)

	var name string
	for i := 0; i < min(len(pp), lvl+1); i++ {
		name = strings.ToLower(replacer.Replace(pp[i])) + name
	}

	return name
}



func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func reverse(a []string) {
	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}
}
