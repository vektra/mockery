// +build custom2

package comment

type IfaceWithCustomBuildTagInComment interface {
	Sprintf(format string, a ...interface{}) string
}
