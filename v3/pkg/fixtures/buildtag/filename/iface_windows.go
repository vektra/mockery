package filename

type IfaceWithBuildTagInFilename interface {
	Sprintf(format string, a ...interface{}) string
}
