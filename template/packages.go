package template

type Packages []*Package

func (p Packages) SyncPkgQualifier() string {
	for _, imprt := range p {
		if imprt.Path() == "sync" {
			return imprt.Qualifier()
		}
	}

	return "sync"
}
