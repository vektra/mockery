# `#!yaml inpackage:`

:octicons-tag-24: v3.5.0

The `#!yaml inpackage` parameter overrides mockery's auto-detection logic that determines whether an output file resides inside or outside of the package of the original interface. Normally, import statements and type qualifiers (such as `srcpkg.TypeName`, the qualified variant of `TypeName`) are added when the mock references types in the original package _and_ the mock resides outside of that package. For example:

```go
import (
	mock "github.com/stretchr/testify/mock"
	"github.com/vektra/mockery/v3/internal/fixtures/inpackage"
)

// Bar provides a mock function for the type MockFoo
func (_mock *MockFoo) Bar() inpackage.InternalStringType {
```

When `#!yaml inpackage: true` is set, the `"github.com/vektra/mockery/v3/internal/fixtures/inpackage"` import is removed and the type will be referred to as its unqualified name. For example:

```go
import (
	mock "github.com/stretchr/testify/mock"
)

// Bar provides a mock function for the type MockFoo
func (_mock *MockFoo) Bar() InternalStringType {
```