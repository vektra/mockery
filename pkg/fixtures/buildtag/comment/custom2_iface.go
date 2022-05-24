//go:build custom2
// +build custom2

package comment

type IfaceWithCustomBuildTagInComment interface {
	Sprintf(format string, a ...interface{}) string
}
