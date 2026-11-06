package argspackagename

import "github.com/vektra/mockery/v3/internal/fixtures/args_package_name/args"

// UsesArgsPackage has methods that take a type from a package named "args".
// The generated mocks must not name the argument of the Run closure "args".
type UsesArgsPackage interface {
	Single(v args.Value)
	Variadic(v args.Value, rest ...args.Value)
}
