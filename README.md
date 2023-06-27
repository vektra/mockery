
mockery
=======
[![Release](https://github.com/vektra/mockery/actions/workflows/release.yml/badge.svg)](https://github.com/vektra/mockery/actions/workflows/release.yml) [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/vektra/mockery/v2?tab=overview) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/vektra/mockery) ![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/vektra/mockery) [![Go Report Card](https://goreportcard.com/badge/github.com/vektra/mockery)](https://goreportcard.com/report/github.com/vektra/mockery) [![codecov](https://codecov.io/gh/vektra/mockery/branch/master/graph/badge.svg)](https://codecov.io/gh/vektra/mockery)

mockery provides the ability to easily generate mocks for Golang interfaces using the [stretchr/testify/mock](https://pkg.go.dev/github.com/stretchr/testify/mock?tab=doc) package. It removes the boilerplate coding required to use mocks.

Documentation
--------------

Documentation is found at out [GitHub Pages site](https://vektra.github.io/mockery/).

Development
------------

taskfile.dev is used for build tasks. Initialize all go build tools:

```
go mod download -x
```

You can run any of the steps listed in `Taskfile.yml`:

```
$ task test
task: [test] go test -v -coverprofile=coverage.txt ./...
```

Development Efforts
-------------------

### v1

v1 is the original version of the software, and is no longer supported.

### v2

`mockery` is currently in v2, which originally included cosmetic and configuration improvements over v1, but also implements a number of quality-of-life additions.

### v3

[v3](https://github.com/vektra/mockery/projects/3) will include a ground-up overhaul of the entire codebase and will completely change how mockery works internally and externally. The highlights of the project are:
- Moving towards a package-based model instead of a file-based model. `mockery` currently iterates over every file in a project and calls `package.Load` on each one, which is time-consuming. Moving towards a model where the entire package is loaded at once will dramatically reduce runtime, and will simplify logic. Additionally, supporting only a single mode of operation (package mode) will greatly increase the intuitiveness of the software.
- Configuration-driven generation. `v3` will be entirely driven by configuration, meaning:
  * You specify the packages you want mocked, instead of relying on it auto-discovering your package. Auto-discovery in theory sounds great, but in practice it leads to a great amount of complexity for very little benefit.
  * Package- or interface-specific overrides can be given that change mock generation settings on a granular level. This will allow your mocks to be generated in a heterogeneous manner, and will be made explicit by YAML configuration.
 - Proper error reporting. Errors across the board will be done in accordance with modern Golang practices
 - Variables in generated mocks will be given meaningful names.



Stargazers
----------

[![Stargazers over time](https://starchart.cc/vektra/mockery.svg)](https://starchart.cc/vektra/mockery)
