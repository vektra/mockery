Additional Notes
================

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