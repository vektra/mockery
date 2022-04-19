package foo

import "github.com/vektra/mockery/v2/pkg/fixtures/example_project/bar/foo"

type PackageNameSameAsImport interface {
	NewClient() foo.Client
}
