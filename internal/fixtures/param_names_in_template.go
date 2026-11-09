package test

// ParamNamesInTemplate has parameters named after identifiers that the
// matryer template uses inside each generated method.
type ParamNamesInTemplate interface {
	Set(mock string, callInfo int) error
}
