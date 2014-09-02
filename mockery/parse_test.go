package mockery

import (
	"os"
	"path/filepath"
	"testing"

	"go/ast"

	"github.com/stretchr/testify/assert"
)

var fixturePath string
var testFile string
var testFile2 string

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	fixturePath = filepath.Join(dir, "fixtures")

	testFile = filepath.Join(dir, "fixtures", "requester.go")
	testFile2 = filepath.Join(dir, "fixtures", "requester2.go")
}

func TestFileParse(t *testing.T) {
	parser := NewParser()

	err := parser.Parse(testFile)
	assert.NoError(t, err)

	node, err := parser.Find("Requester")
	assert.NoError(t, err)
	assert.NotNil(t, node)

	_, ok := node.(*ast.InterfaceType)

	assert.True(t, ok)
}
