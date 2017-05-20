// +build linux

package comment

type IfaceWithBuildTagInComment interface {
	Sprintf(format string, a ...interface{}) string
}
