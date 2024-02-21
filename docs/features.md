Features
========

Mock Styles
------------

:octicons-tag-24: v2.42.0 :material-test-tube:{ .tooltip foo }

!!! warning "ALPHA"
    This feature is currently in alpha. Usage is not recommended for production applications as the configuration semantics and/or mock behavior might change in future releases.

mockery provides a parameter called `#!yaml style:` that allows you to specify different styles of mock implementations. This parameter defaults to `#!yaml style: mockery`, but you may also provide other values described below:

| style | description |
|-------|-------------|
| `mockery` | The default style. Generates the mocks that you typically expect from this project. |
| `moq` | Generates mocks in the style of [Mat Ryer's project](https://github.com/matryer/moq). |

This table describes the characteristics of each mock style.

| feature | `mockery` | `moq` |
|---------|-----------|-------|
| Uses testify | :white_check_mark: | :x: |
| Strictly typed arguments | :warning:[^1] | :white_check_mark: |
| Stricly typed return values | :white_check_mark: | :white_check_mark: |
| Multiple mocks per file | :x: | :x: |
| Return values mapped from argument values | :white_check_mark: | :x: |
| Supports generics | :white_check_mark: | :white_check_mark: |
| 

[^1]: 
	If you use the `.RunAndReturn` method in the [`.EXPECT()`](/mockery/features/#expecter-structs) structs, you can get strict type safety. Otherwise, `.EXPECT()` methods are type-safe-ish, meaning you'll get compiler errors if you have the incorrect number of arguments, but the types of the arguments themselves are `interface{}` due to the need to specify `mock.Anything` in some cases.

### `mockery`

`mockery` is the default style of mocks you are used to with this project.

### `moq`

The `moq` implementations are generated using similar logic as the project at https://github.com/matryer/moq. The mocks rely on you defining stub functions that will be called directly from each method. For example, assume you have an interface like this:

```go
type Requester interface {
	Get(path string) (string, error)
}
```

`moq` will create an implementation that looks roughly like this:

```go
type RequesterMock struct {
	// GetFunc mocks the Get method.
	GetFunc func(path string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Get holds details about calls to the Get method.
		Get []struct {
			// Path is the path argument value.
			Path string
		}
	}
	lockGet sync.RWMutex
}

// Get calls GetFunc.
func (mock *RequesterMock) Get(path string) (string, error) {
	if mock.GetFunc == nil {
		panic("RequesterMock.GetFunc: method is nil but Requester.Get was just called")
	}
	callInfo := struct {
		Path string
	}{
		Path: path,
	}
	mock.lockGet.Lock()
	mock.calls.Get = append(mock.calls.Get, callInfo)
	mock.lockGet.Unlock()
	return mock.GetFunc(path)
}
```

The way you interact with this mock in a test would look like this:

```go
func TestRequester(t *testing.T){
	requesterMock := &RequesterMock{
		GetFunc: func(path string) (string, error) {
			return "value", fmt.Errorf("failed")
		}
	}
}
```

#### Configuration

A simple basic configuration might look like this:

```yaml
style: moq
packages:
  github.com/vektra/mockery/v2/pkg/fixtures:
    config:
      template-map:
        with-resets: true
        skip-ensure: true
        stub-impl: false
```

The `moq` style is driven entirely off of Go templates. Various features of the moq may be toggled by using the `#! template-map:` mapping. These are the moq-specific parameters:

| name | default | description |
|------|---------|-------------|
| `#!yaml with-resets` | `<no value>` | Generates functions that aid in resetting the list of calls made to a mock. |
| `#!yaml skip-ensure` | `<no value>` | Skips adding a compile-time check that asserts the generated mock properly implements the interface. |
| `#!yaml stub-impl` | `<no value>` | Sets methods to return the zero value when no mock implementation is provided. |

!!! note
	Many of the parameters listed in the [parameter descriptions page](/mockery/configuration/#parameter-descriptions) also apply to `moq`.  However, some parameters of `mockery`-styled mocks will not apply to `moq`. Efforts will be made to ensure the documentation is clear on which parameters apply to which style.

Replace Types
-------------

:octicons-tag-24: v2.23.0

The `replace-type` parameter allows adding a list of type replacements to be made in package and/or type names.
This can help overcome issues like usage of type aliases that point to internal packages.

The format of the parameter is:


`originalPackagePath.originalTypeName=newPackageName:newPackagePath.newTypeName`


For example:

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

We can use `replace-type` with only the package part to replace any import of `cloud.google.com/go/internal/pubsub` to
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

Generic type constraints can also be replaced by targeting the changed parameter with the square bracket notation on the left hand side.

```shell
mockery --replace-type github.com/vektra/mockery/v2/baz/internal/foo.InternalBaz[T]=github.com/vektra/mockery/v2/baz.Baz
```

For example:

```go
type InternalBaz[T any] struct{}

func (*InternalBaz[T]) Foo() T {}

// Becomes
type InternalBaz[T baz.Baz] struct{}

func (*InternalBaz[T]) Foo() T {}
```

If a type constraint needs to be removed and replaced with a type, target the constraint with square brackets and include a '-' in front to have it removed.

```shell
mockery --replace-type github.com/vektra/mockery/v2/baz/internal/foo.InternalBaz[-T]=github.com/vektra/mockery/v2/baz.Baz
```

For example:

```go
type InternalBaz[T any] struct{}

func (*InternalBaz[T]) Foo() T {}

// Becomes
type InternalBaz struct{}

func (*InternalBaz) Foo() baz.Baz {}
```

When replacing a generic constraint, you can replace the type with a pointer by adding a '*' before the output type name.

```shell
mockery --replace-type github.com/vektra/mockery/v2/baz/internal/foo.InternalBaz[-T]=github.com/vektra/mockery/v2/baz.*Baz
```

For example:

```go
type InternalBaz[T any] struct{}

func (*InternalBaz[T]) Foo() T {}

// Becomes
type InternalBaz struct{}

func (*InternalBaz) Foo() *baz.Baz {}
```

`packages` configuration
------------------------
:octicons-tag-24: v2.21.0

!!! info
    See the [Migration Docs](migrating_to_packages.md) on how to migrate to this new feature.

Mockery has a configuration parameter called `packages`. In this config section, you define the packages and the interfaces you want mocks generated for. The packages can be any arbitrary package, either your own project or anything within the Go ecosystem. You may provide package-level or interface-level overrides to the default config you provide.

Usage of the `packages` config section is desirable for multiple reasons:

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
          - mockname: MockRequesterVariadicOneArgument
            unroll-variadic: False
          - mockname: MockRequesterVariadic
  io:
    config:
      all: True # (3)!
    interfaces:
      Writer:
        config:
          with-expecter: False # (4)!
```

1.  For this package, we provide no package-level config (which means we inherit the defaults at the top-level). Since our default of `all:` is `False`, mockery will only generate the interfaces we specify. We tell it which interface to generate by using the `interfaces` section and specifying an empty map, one for each interface.
2. There might be cases where you want multiple mocks generated from the same interface. To do this, you can define a default `config` section for the interface, and further `configs` (plural) section, one for each mock. You _must_ specify a `mockname` for the mocks in this section to differentiate them.
3. This is telling mockery to generate _all_ interfaces in the `io` package.
4. We can provide interface-specific overrides to the generation config.

### Templated variables

!!! note
    Templated variables are only available when using the `packages` config feature.

Included with this feature is the ability to use templated strings for various configuration options. This is useful to define where your mocks are placed and how to name them. You can view the template variables available in the [Configuration](configuration.md#template-variables) section of the docs.

### Recursive package discovery

:octicons-tag-24: v2.25.0

When `#!yaml recursive: true` is set on a particular package:

```yaml
packages:
  github.com/user/project:
    config:
      recursive: true
      with-expecter: true
```

mockery will dynamically discover all sub-packages within the specified package. This is done by calling `packages.Load` on the specified package, which induces Go to download the package from the internet (or simply your local project). Mockery then recursively discovers all sub-directories from the root package that also contain `.go` files and injects the respective package path into the config map as if you had specified them manually. As an example, your in-memory config map may end up looking like this:

```yaml
packages:
  github.com/user/project:
    config:
      recursive: true
      with-expecter: true
  github.com/user/project/subpkg1:
    config:
      recursive: true
      with-expecter: true
  github.com/user/project/subpkg2:
    config:
      recursive: true
      with-expecter: true
```

You can use the `showconfig` command to see the config mockery injects. The output of `showconfig` theoretically could be copy-pasted into your yaml file as it is semantically equivalent.

mockery will _not_ recurse into submodules, i.e. any subdirectory that contains a go.mod file. You must specify the submodule as a separate line item in the config if you would like mocks generated for it as well.

??? note "performance characteristics"
    The performance when using `#!yaml recursive: true` may be worse than manually specifying all packages statically in the yaml file. This is because of the fact that mockery has to recursively walk the filesystem path that contains the package in question. It may unnecessarily walk down unrelated paths (for example, a Python virtual environment that is in the same path as your package). For this reason, it is recommended _not_ to use `#!yaml recursive: true` if it can be avoided.

### Regex matching

You can filter matched interfaces using the `include-regex` option. To generate mocks only for interfaces ending in `Client` we can use the following configuration:

```yaml
packages:
  github.com/user/project:
    config:
      recursive: true
      include-regex: ".*Client"
```

To further refine matched interfaces, you can also use `exclude-regex`. If an interface matches both `include-regex` and `exclude-regex` then it will not be generated. For example, to generate all interfaces except those ending in `Func`:

```yaml
packages:
  github.com/user/project:
    config:
      recursive: true
      include-regex: ".*"
      exclude-regex: ".*Func"
```

You can only use `exclude-regex` with `include-regex`. If set by itself, `exclude-regex` has no effect.

??? note "all: true"
    Using `all: true` will override `include-regex` (and `exclude-regex`) and issue a warning.

Mock Constructors
-----------------

:octicons-tag-24: v2.11.0

All mock objects have constructor functions. These constructors do basic test setup so that the expectations you set in the code are asserted before the test exits.

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

:octicons-tag-24: v2.10.0 · `with-expecter: True`

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

:octicons-tag-24: v2.20.0

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
