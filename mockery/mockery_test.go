package mockery

import (
	"os"
	"path/filepath"
)

var fixturePath string

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fixturePath = filepath.Join(dir, "fixtures")
}

// getFixturePath returns an absolute path to a fixture sub-directory or file.
//
// getFixturePath("src.go") returns "/path/to/fixtures/src.go"
// getFixturePath("a", "b", "c", "src.go") returns "/path/to/fixtures/a/b/c/src.go"
func getFixturePath(subdirOrBasename ...string) string {
	return filepath.Join(append([]string{fixturePath}, subdirOrBasename...)...)
}
