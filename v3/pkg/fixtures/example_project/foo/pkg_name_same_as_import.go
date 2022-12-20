package foo

import "github.com/vektra/mockery/v3/pkg/fixtures/example_project/bar/foo"

type PackageNameSameAsImport interface {
	NewClient() foo.Client
}
