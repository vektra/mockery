package template

type Mocks []MockData

// MocksSomeMethod returns true of any one of the Mocks has at least 1
// method.
func (m Mocks) MocksSomeMethod() bool {
	for _, mock := range m {
		if len(mock.Methods) > 0 {
			return true
		}
	}

	return false
}
