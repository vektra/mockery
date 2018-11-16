package mockery

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var testFile string
var testFile2 string

func init() {
	testFile = getFixturePath("requester.go")
	testFile2 = getFixturePath("requester2.go")
}

func TestFileParse(t *testing.T) {
	parser := NewParser(nil)

	err := parser.Parse(testFile)
	assert.NoError(t, err)

	err = parser.Load()
	assert.NoError(t, err)

	node, err := parser.Find("Requester")
	assert.NoError(t, err)
	assert.NotNil(t, node)
}

func noTestFileInterfaces(t *testing.T) {
	parser := NewParser(nil)

	err := parser.Parse(testFile)
	assert.NoError(t, err)

	err = parser.Load()
	assert.NoError(t, err)

	nodes := parser.Interfaces()
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, "Requester", nodes[0].Name)
}

func TestBuildTagInFilename(t *testing.T) {
	parser := NewParser(nil)

	// Include the major OS values found on https://golang.org/dl/ so we're likely to match
	// anywhere the test is executed.
	err := parser.Parse(getFixturePath("buildtag", "filename", "iface_windows.go"))
	assert.NoError(t, err)
	err = parser.Parse(getFixturePath("buildtag", "filename", "iface_linux.go"))
	assert.NoError(t, err)
	err = parser.Parse(getFixturePath("buildtag", "filename", "iface_darwin.go"))
	assert.NoError(t, err)
	err = parser.Parse(getFixturePath("buildtag", "filename", "iface_freebsd.go"))
	assert.NoError(t, err)

	err = parser.Load()
	assert.NoError(t, err) // Expect "redeclared in this block" if tags aren't respected

	nodes := parser.Interfaces()
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, "IfaceWithBuildTagInFilename", nodes[0].Name)
}

func TestBuildTagInComment(t *testing.T) {
	parser := NewParser(nil)

	// Include the major OS values found on https://golang.org/dl/ so we're likely to match
	// anywhere the test is executed.
	err := parser.Parse(getFixturePath("buildtag", "comment", "windows_iface.go"))
	assert.NoError(t, err)
	err = parser.Parse(getFixturePath("buildtag", "comment", "linux_iface.go"))
	assert.NoError(t, err)
	err = parser.Parse(getFixturePath("buildtag", "comment", "darwin_iface.go"))
	assert.NoError(t, err)
	err = parser.Parse(getFixturePath("buildtag", "comment", "freebsd_iface.go"))
	assert.NoError(t, err)

	err = parser.Load()
	assert.NoError(t, err) // Expect "redeclared in this block" if tags aren't respected

	nodes := parser.Interfaces()
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, "IfaceWithBuildTagInComment", nodes[0].Name)
}

func TestCustomBuildTag(t *testing.T) {
	parser := NewParser([]string{"custom"})

	// Include two files that define the same interface, but with different
	// build tags. Only one should be loaded.
	err := parser.Parse(getFixturePath("buildtag", "comment", "custom_iface.go"))
	assert.NoError(t, err)
	err = parser.Parse(getFixturePath("buildtag", "comment", "custom2_iface.go"))
	assert.NoError(t, err)

	err = parser.Load()
	assert.NoError(t, err) // Expect "redeclared in this block" if tags aren't respected

	found := false
	for _, node := range parser.Interfaces() {
		if node.Name == "IfaceWithCustomBuildTagInComment" {
			found = true
			break
		}
	}
	assert.True(t, found, "IfaceWithCustomBuildTagInComment not parsed")
}
