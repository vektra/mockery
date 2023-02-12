Additional Notes
================

Variadic Arguments
------------------

When mocking methods with variadic arguments, some complexities are introduced. Before this PR: https://github.com/vektra/mockery/pull/123, mocking a variadic method looked like this:

```go
type Foo interface {
  Bar(s ...string) error
}

func TestFoo(t *testing.T) {
  m := NewMockFoo(t)
  m.On("Bar", []string{"hello", "world"}).Return(nil)
}
```

After the PR, you could use this syntax:

```go
func TestFoo(t *testing.T) {
  m := NewMockFoo(t)
  m.On("Bar", "hello", "world").Return(nil)
```

This introduces ambiguities because if you want to do something like this:

```
m.On("Bar", mock.Anything).Return(nil)
```

This is impossible to distinguish between these two intentions:
1. Any number of variadic arguments of any value
2. A single variadic argument of any value

This is fixed in https://github.com/vektra/mockery/pull/359 where you can provide `unroll-variadic: False` to get back to the old behavior. Thus, if you want to assert the first case, you can then do:

```
m.On("Bar", mock.Anything).Return(nil)
```

If you want to specify the second case, you must set `unroll-variadic: True`. Then this assertion's intention will be modified to mean the second case:

```
m.On("Bar", mock.Anything).Return(nil)
```

An upstream patch to `testify` is currently underway to allow passing `mock.Anything` directly to the variadic slice: https://github.com/stretchr/testify/pull/1348

If this is merged, it would become possible to describe the above two cases respectively:

```go
// case 1
m.On("Bar", mock.Anything).Return(nil)
// case 2
m.On("Bar", []interface{}{mock.Anything}).Return(nil)
```

References:
- https://github.com/vektra/mockery/pull/359
- https://github.com/vektra/mockery/pull/123
- https://github.com/vektra/mockery/pull/550
- https://github.com/vektra/mockery/issues/541

Semantic Versioning
-------------------

The versioning in this project applies only to the behavior of the mockery binary itself. This project explicitly does not promise a stable internal API, but rather a stable executable. The versioning applies to the following:

1. CLI arguments.
2. Parsing of Golang code. New features in the Golang language will be supported in a backwards-compatible manner, except during major version bumps.
3. Behavior of mock objects. Mock objects can be considered to be part of the public API.
4. Behavior of mockery given a set of arguments.

What the version does _not_ track:
1. The interfaces, objects, methods etc. in the vektra/mockery package.
2. Compatibility of `go get`-ing mockery with new or old versions of Golang.

Mocking interfaces in `main`
----------------------------

When your interfaces are in the main package, you should supply the `--inpackage` flag.
This will generate mocks in the same package as the target code, avoiding import issues.
