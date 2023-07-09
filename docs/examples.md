Examples
========

!!! tip
	IDEs are really useful when interacting with mockery objects. All mockery objects embed the [`github.com/stretchr/testify/mock.Mock`](https://pkg.go.dev/github.com/stretchr/testify/mock#Mock) object so you have access to both methods provided by mockery, and from testify itself. IDE auto-completion will show you all methods available for your use.

### Simple case

Given this interface:

```go title="string.go"
package example_project

type Stringer interface {
	String() string
}
```

Create a mock for this interface by specifying it in your config. We can then create a test using this new mock object:

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

Note that in combination with using the mock's constructor and the [`.EXPECT()`](features.md#expecter-structs) directives, your test will automatically fail if the expected call is not made. 

??? tip "Alternate way of specifying expectations"

	You can also use the `github.com/stretchr/testify/mock.Mock` object directly (instead of using the `.EXPECT()` methods, which provide type-safe-ish assertions).

	```go title="string_test.go"
	func TestString(t *testing.T) {
		mockStringer := NewMockStringer(t)
		mockStringer.On("String").Return("mockery")
		assert.Equal(t, "mockery", Foo(mockStringer))
	}
	```

	We recommend always interacting with the assertions through `.EXPECT()` as mockery auto-generates methods that call out to `Mock.On()` themselves, providing you with some amount of compile-time safety. Consider if all your expectations for `String()` use the `Mock.On()` methods, and you decide to add an argument to `String()` to become `String(foo string)`. Now, your existing tests will only fail when you run them. If you had used `.EXPECT()` and regenerated your mocks after changing the function signature, your IDE, and the go compiler itself, would both tell you immediately that your expectations don't match the function signature. 

### Function type case

!!! bug
	Generating mocks for function types is likely not functioning in the `packages` config semantics. You'll likely need to revert to the legacy semantics as shown below.

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
