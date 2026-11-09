package test

// ReturnFuncParamNames has parameters named after identifiers that the
// testify template declares when it looks up a RunAndReturn function.
type ReturnFuncParamNames interface {
	Set(key string, ok bool) (int, error)
	Enabled(ok bool) bool
	Apply(returnFunc func() int) int
}
