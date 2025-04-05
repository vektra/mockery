package template

type Interfaces []Interface

// ImplementsSomeMethod returns true if any one of the Mocks has at least 1 method.
func (m Interfaces) ImplementsSomeMethod() bool {
	for _, mock := range m {
		if len(mock.Methods) > 0 {
			return true
		}
	}

	return false
}
