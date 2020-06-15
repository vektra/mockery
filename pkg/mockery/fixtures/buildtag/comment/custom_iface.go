// +build custom

package comment

type IfaceWithCustomBuildTagInComment interface {
	Sprintf(format string, a ...interface{}) string
}
