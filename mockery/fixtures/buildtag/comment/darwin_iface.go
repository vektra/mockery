// +build darwin

package comment

type IfaceWithBuildTagInComment interface {
	Sprintf(format string, a ...interface{}) string
}
