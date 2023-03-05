Features
========

`packages` configuration
------------------------
:octicons-tag-24: 2.21.0 · :material-test-tube: Alpha Feature

[Github Discussion](https://github.com/vektra/mockery/discussions/549)

Mockery has a configuration parameter called `packages`. This config represents a huge paradigm shift that is highly recommended for the large amount of flexibility it grants you.

In this config section, you define the packages and the intefaces you want mocks generated for. The packages can be any arbitrary package, either your own project or anything within the Go ecosystem. You may provide package-level or interface-level overrides to the default config you provide.

Usage of the `packages` config section is desirable for mutiple reasons:

1. Up to 5x increase in mock generation speed over the legacy method
2. Granular control over interface generation, location, and file names
3. Singular location for all config, instead of spread around by `//go:generate` statements
4. Clean, easy to understand.

### Examples

Here is an example configuration set:

```yaml
with-expecter: True
packages:
  github.com/vektra/mockery/v2/pkg: # (1)!
    interfaces:
      TypesPackage:
      RequesterVariadic:
        config: # (2)!
          with-expecter: False 
        configs:
          - structname: RequesterVariadicOneArgument
            unroll-variadic: False
          - structname: RequesterVariadic
  io:
    config:
      all: True # (3)!
    interfaces:
      Writer:
        config:
          with-expecter: False # (4)!
```

1.  For this package, we provide no package-level config (which means we inherit the deafults at the top-level). Since our default of `all:` is `False`, mockery will only generate the interfaces we specify. We tell it which interface to generate by using the `interfaces` section and specifying an empty map, one for each interface.
2. There might be cases where you want multiple mocks generated from the same interface. To do this, you can define a default `config` section for the interface, and further `configs` (plural) section, one for each mock. You _must_ specify a `structname` for the mocks in this section to differentiate them.
3. This is telling mockery to generate _all_ interfaces in the `io` package.
4. We can provide interface-specifc overrides to the generation config.

### Templated directory and filenames

Included with this feature is the ability to use templated strings for the destination directory and filenames of the generated mocks.

The default parameters are:

```yaml title="Defaults"
filename: "mock_{{.InterfaceName}}.go"
dir: "mocks/{{.PackagePath}}"
```

The template variables available for your use are:

| name | description |
|------|-------------|
| InterfaceName | The name of the original interface being mocked |
| PackageName | The name of the package from the original interface |
| Package Path | The fully qualified package path of the original interface |
| MockName | The name of the generated mock |

Mock Constructors
-----------------

:octicons-tag-24: 2.11.0

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

:octicons-tag-24: 2.10.0 · `with-expecter: True`

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

:octicons-tag-24: 2.20.0

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
