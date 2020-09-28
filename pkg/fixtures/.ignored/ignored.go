package ignored

// TestInterface should not be mock'ed as it is part of a directory that has a '.' prefix.
type TestInterface interface {
	Ignored() bool
}
