//go:build freebsd
// +build freebsd

package comment

type IfaceWithBuildTagInComment interface {
	Sprintf(format string, a ...interface{}) string
}
