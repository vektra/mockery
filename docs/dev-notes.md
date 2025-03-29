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



Go Syntax Updates
------------------

When Go releases new syntax, there are two approaches that the mockery project will take:

### Mockery does not need to interact with the new syntax

In such cases, the mockery project _only_ needs to upgrade its `golang.org/x/tools` dependency. This is necessary for the parsing step to simply not fail if it encounters new syntax. If mockery does not need to interact or understand this syntax, this dependency upgrade is likely all that's needed.

Take for example the problems mentioned [here] when mockery upgraded to `go 1.24` in its `go.mod` file. In this situation, the project maintainers wanted to allow mockery to not crash when it parses syntax containing generic type alias syntax. However, the project did not _parse_ this syntax, so the only thing we needed to do was upgrade `golang.org/x/tools`.

### Mockery _does_ need to interact with the new syntax

This situation was encountered in Go 1.18 when generics were introduced. [In this case](https://github.com/vektra/mockery/pull/456/files#diff-33ef32bf6c23acb95f5902d7097b7a1d5128ca061167ec0716715b0b9eeaa5f6), the project needed to be upgraded to `go 1.18` because mockery now had to directly parse and interpret generic types through the `go/ast` package. This was needed in conjunction with an upgrade of `golang.org/x/tools` that handles the actual parsing into `go/ast` data.

It's possible in future versions of Go that only the `toolchain` directive needs to be upgraded to allow mockery to use a more recent `go/ast` package. The purposes of the `go` directive is supposed to inform the compiler what features of the Go language the module uses in its own syntax, so in this case as long as mockery does not itself use generics, it can parse generic type information from _other_ projects without needing to pin its `go` directive to the relevant version.
