Features
========

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
requesterMock.EXPECT().
	Get(mock.Anything).
	Run(func(path string) { fmt.Println(path, "was called") }).
	// Can still use return functions by getting the embedded mock.Call
	Call.Return(func(path string) string { return "result for " + path }, nil)
```

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
