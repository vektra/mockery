package pkg

import (
	"path/filepath"
	"runtime"
)

var rootPath, testFile, testFile2 string

func init() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to determine current file path")
	}

	rootPath = filepath.Dir(filepath.Dir(file))

	testFile = getFixturePath("requester.go")
	testFile2 = getFixturePath("requester2.go")
}

// getFixturePath returns an absolute path to a fixture sub-directory or file.
//
// getFixturePath("src.go") returns "/path/to/pkg/fixtures/src.go"
// getFixturePath("a", "b", "c", "src.go") returns "/path/to/pkg/fixtures/a/b/c/src.go"
func getFixturePath(subdirOrBasename ...string) string {
	return filepath.Join(append([]string{rootPath, "pkg", "fixtures"}, subdirOrBasename...)...)
}

// getMocksPath returns an absolute path to a mocks sub-directory or file.
//
// getMocksPath("src.go") returns "/path/to/mocks/pkg/fixtures/src.go"
// getMocksPath("a", "b", "c", "src.go") returns "/path/to/mocks/pkg/fixtures/a/b/c/src.go"
func getMocksPath(subdirOrBasename ...string) string {
	return filepath.Join(append([]string{rootPath, "mocks", "pkg", "fixtures"}, subdirOrBasename...)...)
}
