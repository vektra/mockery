Features
========

Mock Constructors
-----------------

All mock objects have constructor functions. These constructors do basic test setup so that the expectations you set in the code are asserted before the test exist.

Previously something like this would need to be done:
```go
factory := &mocks.Factory{}
factory.Test(t) // so that mock does not panic when a method is unexpected
defer factory.AssertExpectations(t)
```

Instead, you may simply use the constructor:
```go
factory := mocks.NewFactory(t)
```

The constructor sets up common functionalities automatically
- The `AssertExpectations` method is registered to be called at the end of the tests via `t.Cleanup()` method.
- The testing.TB interface is registered on the `mock.Mock` so that tests don't panic when a call on the mock is unexpected.


Expecter Structs
----------------

|config|`with-expecter: True`|
|-|-|

Mockery now supports an "expecter" struct, which allows your tests to use type-safe methods to generate call expectations. When enabled through the `with-expecter: True` mockery configuration, you can enter into the expecter interface by simply calling `.EXPECT()` on your mock object.

For example, given an interface such as
```go
type Requester interface {
	Get(path string) (string, error)
}
```

You can use the expecter interface as such:
```go
requesterMock := mocks.NewRequester(t)
requesterMock.EXPECT().Get("some path").Return("result", nil)
```

A `RunAndReturn` method is also available on the expecter struct that allows you to dynamically set a return value based on the input to the mock's call.

```go
requesterMock.EXPECT().
	Get(mock.Anything).
	RunAndReturn(func(path string) string { 
		fmt.Println(path, "was called")
		return "result for " + path
	})
```

!!! note 

	Note that the types of the arguments on the `EXPECT` methods are `interface{}`, not the actual type of your interface. The reason for this is that you may want to pass `mock.Any` as an argument, which means that the argument you pass may be an arbitrary type. The types are still provided in the expecter method docstrings.


Return Value Providers
----------------------

Return Value Providers can be used one of two ways.  You may either define a single function with the exact same signature (number and type of input and return parameters) and pass that as a single value to `Return`, or you may pass multiple values to `Return` (one for each return parameter of the mocked function.)  If you are using the second form, for each of the return values of the mocked function, `Return` needs a function which takes the same arguments as the mocked function, and returns one of the return values. For example, if the return argument signature of `passthrough` in the above example was instead `(string, error)` in the interface, `Return` would also need a second function argument to define the error value:

```go
type Proxy interface {
passthrough(ctx context.Context, s string) (string, error)
}
```

First form:

```go
proxyMock := mocks.NewProxy(t)
proxyMock.On("passthrough", mock.AnythingOfType("context.Context"), mock.AnythingOfType("string")).
Return(
    func(ctx context.Context, s string) (string, error) {
        return s, nil
    }
)
```


Second form:

```go
proxyMock := mocks.NewProxy(t)
proxyMock.On("passthrough", mock.AnythingOfType("context.Context"), mock.AnythingOfType("string")).
Return(
    func(ctx context.Context, s string) string {
        return s
    },
    func(ctx context.Context, s string) error {
        return nil
    },
)
```

Replace Types
-------------

The `replace-type` parameter allows adding a list of type replacements to be made in package and/or type names.
This can help overcome some parsing problems like type aliases that the Go parser doesn't provide enough information.

This parameter can be specified multiple times.

```shell
mockery --replace-type github.com/vektra/mockery/v2/baz/internal/foo.InternalBaz=baz:github.com/vektra/mockery/v2/baz.Baz
```

This will replace any imported named `"github.com/vektra/mockery/v2/baz/internal/foo"`
with `baz "github.com/vektra/mockery/v2/baz"`. The alias is defined with `:` before
the package name. Also, the `InternalBaz` type that comes from this package will be renamed to `baz.Baz`.

This next example fixes a common problem of type aliases that point to an internal package.

`cloud.google.com/go/pubsub.Message` is a type alias defined like this:

```go
import (
    ipubsub "cloud.google.com/go/internal/pubsub"
)

type Message = ipubsub.Message
```

The Go parser that mockery uses doesn't provide a way to detect this alias and sends the application the package and
type name of the type in the internal package, which will not work.

We can use "replace-type" with only the package part to replace any import of `cloud.google.com/go/internal/pubsub` to
`cloud.google.com/go/pubsub`. We don't need to change the alias or type name in this case, because they are `pubsub`
and `Message` in both cases.

```shell
mockery --replace-type cloud.google.com/go/internal/pubsub=cloud.google.com/go/pubsub
```

Original source:

```go
import (
    "cloud.google.com/go/pubsub"
)

type Handler struct {
    HandleMessage(m pubsub.Message) error
}
```

Invalid mock generated without this parameter (points to an `internal` folder):

```go
import (
    mock "github.com/stretchr/testify/mock"

    pubsub "cloud.google.com/go/internal/pubsub"
)

func (_m *Handler) HandleMessage(m pubsub.Message) error {
    // ...
    return nil
}
```

Correct mock generated with this parameter.

```go
import (
    mock "github.com/stretchr/testify/mock"

    pubsub "cloud.google.com/go/pubsub"
)

func (_m *Handler) HandleMessage(m pubsub.Message) error {
    // ...
    return nil
}
```
