Frequently Asked Questions
===========================

error: `no go files found in root search path`
---------------------------------------------

When using the `packages` feature, `recursive: true` and you have specified a package that contains no `*.go` files, mockery is unable to determine the on-disk location of the package in order to continue the recursive package search. This appears to be a limitation of the [golang.org/x/tools/go/packages](https://pkg.go.dev/golang.org/x/tools/go/packages) package that is used to parse package metadata.

The solution is to create a `.go` file in the package's path and add a `package [name]` directive at the top. It doesn't matter what the file is called. This allows mockery to properly read package metadata.

[Discussion](https://github.com/vektra/mockery/discussions/636)

internal error: package without types was imported
---------------------------------------------------

[https://github.com/vektra/mockery/issues/475](https://github.com/vektra/mockery/issues/475)

This issue indicates that you have attempted to use package in your dependency tree (whether direct or indirect) that uses Go language semantics that your currently-running Go version does not support. The solution:

1. Update to the latest go version
2. Delete all cached packages with `go clean -modcache`
3. Reinstall mockery

Additionally, this issue only happens when compiling mockery from source, such as with `go install`. Our docs [recommend not to use `go install`](../installation#go-install) as the success of your build depends on the compatibility of your Go version with the semantics in use. You would not encounter this issue if using one of the installation methods that install pre-built binaries, like downloading the `.tar.gz` binaries, or through `brew install`.

Multiple Expectations With Identical Arguments
-----------------------------------------------

There might be instances where you want a mock to return different values on successive calls that provide the same arguments. For example, we might want to test this behavior:

```go
// Return "foo" on the first call
getter := NewGetter()
assert(t, "foo", getter.Get("key"))

// Return "bar" on the second call
assert(t, "bar", getter.Get("key"))
```

This can be done by using the `.Once()` method on the mock call expectation:

```go
mockGetter := NewMockGetter(t)
mockGetter.EXPECT().Get(mock.anything).Return("foo").Once()
mockGetter.EXPECT().Get(mock.anything).Return("bar").Once()
```

Or you can identify an arbitrary number of times each value should be returned:

```go
mockGetter := NewMockGetter(t)
mockGetter.EXPECT().Get(mock.anything).Return("foo").Times(4)
mockGetter.EXPECT().Get(mock.anything).Return("bar").Times(2)
```

Note that with proper Go support in your IDE, all the available methods are self-documented in autocompletion help contexts.

Variadic Arguments
------------------

Consider if we have a function `#!go func Bar(message ...string) error`. A typical assertion might look like this:

```go
func TestFoo(t *testing.T) {
  m := NewMockFoo(t)
  m.On("Bar", "hello", "world").Return(nil)
```

We might also want to make an assertion that says "any number of variadic arguments":

```go
m.On("Bar", mock.Anything).Return(nil)
```

However, what we've given to mockery is ambiguous because it is impossible to distinguish between these two intentions:

1. Any number of variadic arguments of any value
2. A single variadic argument of any value

This is fixed in [#359](https://github.com/vektra/mockery/pull/359) where you can provide `unroll-variadic: False` to get back to the old behavior. Thus, if you want to assert (1), you can then do:

```go
m.On("Bar", mock.Anything).Return(nil)
```

If you want to assert (2), you must set `unroll-variadic: True`. Then this assertion's intention will be modified to mean the second case:

```go
m.On("Bar", mock.Anything).Return(nil)
```

An upstream patch to `testify` is currently underway to allow passing `mock.Anything` directly to the variadic slice: [https://github.com/stretchr/testify/pull/1348](https://github.com/stretchr/testify/pull/1348)

If this is merged, it would become possible to describe the above two cases respectively:

```go
// case 1
m.On("Bar", mock.Anything).Return(nil)
// case 2
m.On("Bar", []interface{}{mock.Anything}).Return(nil)
```

References:

- [https://github.com/vektra/mockery/pull/359](https://github.com/vektra/mockery/pull/359)
- [https://github.com/vektra/mockery/pull/123](https://github.com/vektra/mockery/pull/123)
- [https://github.com/vektra/mockery/pull/550](https://github.com/vektra/mockery/pull/550)
- [https://github.com/vektra/mockery/issues/541](https://github.com/vektra/mockery/issues/541)

Semantic Versioning
-------------------

The versioning in this project applies only to the behavior of the mockery binary itself. This project explicitly does not promise a stable internal API, but rather a stable executable. The versioning applies to the following:

1. CLI arguments.
2. Parsing of Go code. New features in the Go language will be supported in a backwards-compatible manner, except during major version bumps.
3. Behavior of mock objects. Mock objects can be considered to be part of the public API.
4. Behavior of mockery given a set of arguments.

What the version does _not_ track:

1. The interfaces, objects, methods etc. in the vektra/mockery package.
2. Compatibility of `go get`-ing mockery with new or old versions of Go.

Mocking interfaces in `main`
----------------------------

When your interfaces are in the main package, you should supply the `--inpackage` flag.
This will generate mocks in the same package as the target code, avoiding import issues.

mockery fails to run when `MOCKERY_VERSION` environment variable is set
------------------------------------------------------------------------

This issue was first highlighted [in this GitHub issue](https://github.com/vektra/mockery/issues/391).

mockery uses the viper package for configuration mapping and parsing. Viper is set to automatically search for all config variables specified in its config struct. One of the config variables is named `version`, which gets mapped to an environment variable called `MOCKERY_VERSION`. If you set this environment variable, mockery attempts to parse it into the `version` bool config.

This is an adverse effect of how our config parsing is set up. The solution is to rename your environment variable to something other than `MOCKERY_VERSION`.
