package template

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

func (p *Package) ImportStatement() string {
	if p.Alias == "" {
		return `"` + p.Path() + `"`
	}
	return p.Alias + ` "` + p.Path() + `"`
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

	return p.pkg.Path()
}
