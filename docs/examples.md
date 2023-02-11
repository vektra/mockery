Examples
========

### Simple case

Given this interface:

```go title="string.go"
package example_project

//go:generate --name Stringer
type Stringer interface {
	String() string
}
```

Run: `go generate` (using the recommended config) and the the file `mock_Stringer_test.go` will be generated. You can now use this mock to create assertions and expectations.

```go title="string_test.go"
package example_project

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Foo(s Stringer) string {
	return s.String()
}

func TestString(t *testing.T) {
	mockStringer := NewMockStringer(t)
	mockStringer.EXPECT().String().Return("mockery")
	assert.Equal(t, "mockery", Foo(mockStringer))
}
```

Note that in combination with using the mock's constructor and the `.EXPECT()` directives, your test will automatically fail if the expected call is not made.

### Function type case

Given this is in `send.go`

```go
package test

type SendFunc func(data string) (int, error)
```

Run: `mockery --name=SendFunc` and the following will be output:

```go title="mock_SendFunc_test.go"
package mocks

import (
	"github.com/stretchr/testify/mock"

	testing "testing"
)

type SendFunc struct {
	mock.Mock
}

func (_m *SendFunc) Execute(data string) (int, error) {
	ret := _m.Called(data)

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(data)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(data)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewSendFunc creates a new instance of SendFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSendFunc(t testing.TB) *SendFunc {
	mock := &SendFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
```

### Return Value Provider Functions

If your tests need access to the arguments to calculate the return values,
set the return value to a function that takes the method's arguments as its own
arguments and returns the return value. For example, given this interface:

```go
package test

type Proxy interface {
  passthrough(ctx context.Context, s string) string
}
```

The argument can be passed through as the return value:

```go
import . "github.com/stretchr/testify/mock"

proxyMock := mocks.NewProxy(t)
proxyMock.On("passthrough", mock.AnythingOfType("context.Context"), mock.AnythingOfType("string")).
	Return(func(ctx context.Context, s string) string {
		return s
	})
```

### Note

For mockery to correctly generate mocks, the command has to be run on a module (i.e. your project has to have a go.mod file)