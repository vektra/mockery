Developer Notes
===============

Go Upgrades
------------

The mockery project supports the most recent TWO stable Go versions. Testing matrices will only run on the two most recent stable Go versions. However, given the [Go backwards-compatibility guarantee](https://go.dev/blog/compat), it's very likely projects built off of older Go 1.x syntax will continue to work in perpetuity.(1)
{ .annotate }

1. The caveat, being noted, is the same as the above linked backwards compatibility guarantee:

    > There are a few qualifications to that. First, compatibility means source compatibility. When you update to a new version of Go, you do have to recompile your code. Second, we can add new APIs, but not in a way that breaks existing code.

    > The end of the document warns, “[It] is impossible to guarantee that no future change will break any program.” Then it lays out a number of reasons why programs might still break.

    > For example, it makes sense that if your program depends on a buggy behavior and we fix the bug, your program will break. But we try very hard to break as little as possible and keep Go boring. There are two main approaches we’ve used so far: API checking and testing.

When a new Go minor version `1.N` is released, make the following changes:

1. Upgrade `golang.org/x/tools` to a version that supports Go `1.N`. This is the
   dependency that mockery uses to load and parse Go packages, and it is the only
   dependency that must be upgraded specifically to parse source accepted by the
   new Go release.
2. Update the `go` directive in the root `go.mod` to Go `1.(N-1)`, the older of
   the two supported Go releases. This records mockery's supported Go version
   floor. **Do not set the root module's `go` directive to the newest Go version
   (`1.N`).** Doing so would unnecessarily prevent users of Go `1.(N-1)` from
   building mockery.
3. Update the CI test matrix so that it tests Go `1.(N-1)` and `1.N`.
4. Update the Go image version in the root `Dockerfile` to `1.N`. Unlike the
   root module's `go` directive, the Docker image should use the latest supported
   Go release so that the distributed image can load and parse Go `1.N` source.
5. Regenerate the mocks and run the full test suite with both supported Go
   versions. Review generated-file diffs: changes in `go/types` representations
   or in the standard library can affect generated identifiers or signatures even
   when mockery's templates did not change.

For example, when Go 1.27 is the latest stable release, mockery supports Go 1.26
and Go 1.27, while the root `go.mod` contains `go 1.26`. The
`golang.org/x/tools` dependency must nevertheless be new enough to understand Go
1.27 source code.




Go Syntax Updates
------------------

When Go releases new syntax, there are two approaches that the mockery project will take:

### Mockery does not need to interact with the new syntax

In such cases, `golang.org/x/tools` is the only implementation dependency that
needs to be upgraded. This is necessary so that package loading and parsing do
not fail when they encounter the new syntax. The root module's `go` directive is
still advanced to `1.(N-1)` as described above, but that change records the
project's support policy; it is not what teaches mockery to parse Go `1.N`
syntax.

For example, when mockery needed to accept source containing generic type alias
syntax, mockery did not need to interpret that syntax itself. Upgrading
`golang.org/x/tools` supplied the required parser support; setting the `go`
directive to that newest Go version was neither necessary nor desirable.

### Mockery _does_ need to interact with the new syntax

This situation was encountered in Go 1.18 when generics were introduced. [In this case](https://github.com/vektra/mockery/pull/456/files#diff-33ef32bf6c23acb95f5902d7097b7a1d5128ca061167ec0716715b0b9eeaa5f6), the project needed to be upgraded to `go 1.18` because mockery now had to directly parse and interpret generic types through the `go/ast` package. This was needed in conjunction with an upgrade of `golang.org/x/tools` that handles the actual parsing into `go/ast` data.

The `go` directive controls the language version used to compile mockery's own
module and states the minimum Go version required by that module. It does not
select the version of `go/ast` or add support for parsing newer source in other
projects. Parser support comes from `golang.org/x/tools` and from the Go
toolchain used to build and run mockery. A `toolchain` directive may suggest a
newer toolchain without raising the module's declared minimum Go version, but it
should only be added when mockery actually needs that behavior.
