mockery
=======

mockery provides the ability to easily generate mocks for golang interfaces. It removes
the boilerplate coding required to use mocks.

### Installation

`go get github.com/vektra/mockery`, then `$GOPATH/bin/mockery`

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

 r0 := ret.Get(0).(string)

 return r0
}
```

### Imports

mockery pulls in all the same imports used in the file that contains the interface so
that package types will work correctly. It then runs the output through the `imports` 
package to remove any unnecessary imports (as they'd result in compile errors).

### Types

mockery should handle all types. If you find it does not, please report the issue.

### All

It's common for a big package to have a lot of interfaces, so mockery provides `-all`.
This option will tell mockery to scan all files under the directory named by `-dir` ("." by default)
and generates mocks for any interfaces it finds.

`-all` was designed to be able to be used automatically in the background if required.

### Output

mockery always generates files with the package `mocks` to keep things clean and simple.
You can control which mocks directory is used by using `-output`, which defaults to `./mocks`.

### Debug

Use `mockery -print` to have the resulting code printed out instead of written to disk.
