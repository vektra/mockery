mockery
=======

mockery provides the ability to easily generate mocks for golang interfaces. It removes
the boilerplate coding required to use mocks.

### Installation

`go get github.com/vektra/mockery/.../`, then `$GOPATH/bin/mockery`

### Example

Given this is in `string.go`

```go
package test

type Stringer interface {
  String() string
}
```

Run: `mockery -name=Stringer` and the following will be output to `mocks/Stringer.go`:

```go
package mocks

import "github.com/stretchr/testify/mock"

type Stringer struct {
 mock.Mock
}

func (m *Stringer) String() string {
	ret := m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0)
	}

	return r0
}
```

### Imports

mockery pulls in all the same imports used in the file that contains the interface so
that package types will work correctly. It then runs the output through the `imports`
package to remove any unnecessary imports (as they'd result in compile errors).

### Types

mockery should handle all types. If you find it does not, please report the issue.

### Return Value Provider Functions

If your tests need access to the arguments to calculate the return values,
set the return value to a function that takes the method's arguments as its own
arguments and returns the return value. For example, given this interface:

```go
package test

type Proxy interface {
  passthrough(s string) string
}
```

The argument can be passed through as the return value:

```go
import . "github.com/stretchr/testify/mock"

Mock.On("passthrough", AnythingOfType("string")).Return(func(s string) string {
    return s
})
```

Note, this approach should be used judiciously, as return values should generally 
not depend on arguments in mocks; however, this approach can be helpful for 
situations like passthroughs or other test-only calculations.

### Name

The `-name` option takes either the name or matching regular expression of interface to generate mock(s) for.

### All

It's common for a big package to have a lot of interfaces, so mockery provides `-all`.
This option will tell mockery to scan all files under the directory named by `-dir` ("." by default)
and generates mocks for any interfaces it finds. This option implies `-recursive=true`.

`-all` was designed to be able to be used automatically in the background if required.

### Recursive

Use the `-recursive` option to search subdirectories for the interface(s).
This option is only compatible with `-name`. The `-all` option implies `-recursive=true`.

### Output

mockery always generates files with the package `mocks` to keep things clean and simple.
You can control which mocks directory is used by using `-output`, which defaults to `./mocks`.

## Caseing

mockery generates files using the caseing of the original interface name.  This
can be modified by specifying `-case=underscore` to format the generated file
name using underscore casing.

### Debug

Use `mockery -print` to have the resulting code printed out instead of written to disk.

### Mocking interfaces in `main`

When your interfaces are in the main package you should supply the `-inpkg` flag.
This will generate mocks in the same package as the target code avoiding import issues.
