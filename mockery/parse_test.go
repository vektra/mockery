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
	parser := NewParser()

	err := parser.Parse(testFile)
	assert.NoError(t, err)

	err = parser.Load()
	assert.NoError(t, err)

	node, err := parser.Find("Requester")
	assert.NoError(t, err)
	assert.NotNil(t, node)
}

func noTestFileInterfaces(t *testing.T) {
	parser := NewParser()

	err := parser.Parse(testFile)
	assert.NoError(t, err)

	err = parser.Load()
	assert.NoError(t, err)

	nodes := parser.Interfaces()
	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, "Requester", nodes[0].Name)
}
