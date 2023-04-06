Features
========

Replace Types
-------------

:octicons-tag-24: 2.23.0

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

`packages` configuration
------------------------
:octicons-tag-24: 2.21.0 · :material-test-tube: Alpha Feature

!!! danger
	This feature is considered alpha. It is likely that bugs exist, and subfeatures may be added/subtracted/modified at any time. Use at your own risk. This warning will be updated as this feature matures.

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

1.  For this package, we provide no package-level config (which means we inherit the deafults at the top-level). Since our default of `all:` is `False`, mockery will only generate the interfaces we specify. We tell it which interface to generate by using the `interfaces` section and specifying an empty map, one for each interface.
2. There might be cases where you want multiple mocks generated from the same interface. To do this, you can define a default `config` section for the interface, and further `configs` (plural) section, one for each mock. You _must_ specify a `mockname` for the mocks in this section to differentiate them.
3. This is telling mockery to generate _all_ interfaces in the `io` package.
4. We can provide interface-specifc overrides to the generation config.

### Templated variables

Included with this feature is the ability to use templated strings for various configuration options. This is useful to define where your mocks are placed and how to name them. You can view the template variables available in the [Configuration](/mockery/configuration/#parameter-descriptions) section of the docs, under the `packages config` tab.
  

### Layouts

Using different configuration parameters, we can deploy our mocks on-disk in various ways. These are some common layouts:

!!! info "layouts"

    === "defaults"

        ```yaml
        filename: "mock_{{.InterfaceName}}.go"
        dir: "mocks/{{.PackagePath}}"
        mockname: "Mock{{.InterfaceName}}"
        outpkg: "{{.PackageName}}"
        ```

        If these variables aren't specified, the above values will be applied to the config options. This strategy places your mocks into a separate `mocks/` directory.

        **Interface Description**

        | name | value |
        |------|-------|
        | `InterfaceName` | `MyDatabase` |
        | `PackagePath` | `github.com/user/project/pkgName` |
        | `PackageName` | `pkgName` |

        **Output**

        The mock will be generated at:

        ```
        mocks/github.com/user/project/pkgName/mock_MyDatabase.go
        ```

        The mock file will look like:

        ```go
        package pkgName

        import mock "github.com/stretchr/testify/mock"

        type MockMyDatabase struct {
          mock.Mock
        }
        ```
    === "adjacent to interface"

        ```yaml
        filename: "mock_{{.InterfaceName}}.go"
        dir: "{{.InterfaceDir}}"
        mockname: "Mock{{.InterfaceName}}"
        outpkg: "{{.PackageName}}"
        inpackage: True
        ```

        Instead of the mocks being generated in a different folder, you may elect to generate the mocks alongside the original interface in your package. This may be the way most people define their configs, as it removes circular import issues that can happen with the default config.

        For example, the mock might be generated along side the original source file like this:

        ```
        ./path/to/pkg/db.go
        ./path/to/pkg/mock_MyDatabase.go
        ```

        **Interface Description**

        | name | value |
        |------|-------|
        | `InterfaceName` | `MyDatabase` |
        | `PackagePath` | `github.com/user/project/path/to/pkg`
        | `PackagePathRelative` | `path/to/pkg` |
        | `PackageName` | `pkgName` |
        | `SourceFile` | `./path/to/pkg/db.go` |

        **Output**

        Mock file will be generated at:

        ```
        ./path/to/pkg/mock_MyDatabase.go
        ```

        The mock file will look like:

        ```go
        package pkgName

        import mock "github.com/stretchr/testify/mock"

        type MockMyDatabase struct {
          mock.Mock
        }
        ```

Mock Constructors
-----------------

:octicons-tag-24: 2.11.0

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
