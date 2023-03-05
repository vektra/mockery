Changelog
=========

This changelog describes major feature additions. Please view the `releases` page for more details on commits and minor changes.

### :octicons-tag-24: [`v2.21.0`](https://github.com/vektra/mockery/releases/tag/v2.21.0): `packages` configuration

In this version we release the `packages` configuration section. This new parameter allows defining specific packages to generate mocks for, while also giving fine-grained control over which interfaces are mocked, where they are located, and how they are configured. Details are provided [here](/mockery/features/#packages-configuration).

Community input is desired before we consider deprecations of dynamic walking (via `#!yaml all: True`): https://github.com/vektra/mockery/discussions/549

### :octicons-tag-24: [`v2.20.0`](https://github.com/vektra/mockery/pull/538): Improved Return Value Functions

Return value functions that return an entire method's return value signature can now be provided.

```go
proxyMock := mocks.NewProxy(t)
proxyMock.On("passthrough", mock.AnythingOfType("context.Context"), mock.AnythingOfType("string")).
Return(
    func(ctx context.Context, s string) (string, error) {
        return s, nil
    }
)
```

You may still use the old way where one function is provided for each return value:

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

### :octicons-tag-24: [`2.19.0`](https://github.com/vektra/mockery/releases/tag/v2.19.0): `inpackage-suffix` option

When `inpackage-suffix` is set to `True`, mock files are suffixed with `_mock` instead of being prefixed with `mock_` for InPackage mocks


### :octicons-tag-24: [`v2.16.0`](https://github.com/vektra/mockery/pull/527): Config Search Path

Mockery will iteratively search every directory from the current working directory up to the root path for a `.mockery.yaml` file, if one is not explicitly provided.

### :octicons-tag-24: [`v2.13.0`](https://github.com/vektra/mockery/pull/456): Generics support

Mocks are now capable of supporting Golang generics.

### :octicons-tag-24: [`v2.11.0`](https://github.com/vektra/mockery/pull/406): Mock constructors

Mockery v2.11 introduces constructors for all mocks. This makes instantiation and mock registration a bit easier and
less error-prone (you won't have to worry about forgetting the `AssertExpectations` method call anymore).

Before v2.11:
```go
factory := &mocks.Factory{}
factory.Test(t) // so that mock does not panic when a method is unexpected
defer factory.AssertExpectations(t)
```

After v2.11:
```go
factory := mocks.NewFactory(t)
```

The constructor sets up common functionalities automatically
- The `AssertExpectations` method is registered to be called at the end of the tests via `t.Cleanup()` method.
- The testing.TB interface is registered on the `mock.Mock` so that tests don't panic when a call on the mock is unexpected.

### :octicons-tag-24: [`v2.10.0`](https://github.com/vektra/mockery/pull/396): Expecter Structs

Mockery now supports an "expecter" struct, which allows your tests to use type-safe methods to generate call expectations. When enabled through the `with-expecter: True` mockery configuration, you can enter into the expecter interface by simply calling `.EXPECT()` on your mock object.

For example, given an interface such as
```go
type Requester interface {
	Get(path string) (string, error)
}
```

You can use the type-safe expecter interface as such:
```go
requesterMock := mocks.NewRequester(t)
requesterMock.EXPECT().Get("some path").Return("result", nil)
requesterMock.EXPECT().
	Get(mock.Anything).
	Run(func(path string) { fmt.Println(path, "was called") }).
	// Can still use return functions by getting the embedded mock.Call
	Call.Return(func(path string) string { return "result for " + path }, nil)
```

### :octicons-tag-24: [`v2.0.0`](https://github.com/vektra/mockery/releases/tag/v2.0.0): Major Update

This is the first major update of mockery. Version 2 brings a handful of improvements to mockery:

- Structured and pretty console logging
- CLI now switches over to sp13/cobra
- Use of viper configuration parsing. You can now use a .mockery.yaml config file in your repository
- Various CI fixes and improvements
