mockery
=======
[![Linux Build Status](https://travis-ci.org/vektra/mockery.svg?branch=master)](https://travis-ci.org/vektra/mockery) [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/vektra/mockery/mockery?tab=doc) [![Go Report Card](https://goreportcard.com/badge/github.com/vektra/mockery)](https://goreportcard.com/report/github.com/vektra/mockery)


mockery provides the ability to easily generate mocks for golang interfaces using the [stretchr/testify/mock](https://pkg.go.dev/github.com/stretchr/testify/mock?tab=doc) package. It removes
the boilerplate coding required to use mocks.

Installation
------------

### Github Release

Visit the [releases page](https://github.com/vektra/mockery/releases) to download one of the pre-built binaries for your platform. 

### Docker

Use the Docker image

```
docker pull vektra/mockery
```

### Homebrew

Install through homebrew

```
brew install vektra/tap/mockery
brew upgrade mockery
```

### go get

Alternatively, you can use the DEPRECATED method of:

```
go get github.com/vektra/mockery/.../
``` 

to get a development version of the software.

The SemVer variable by default is `1.0.0` and is overwritten during build steps to the git tagged version. For backwards-compatibility reasons, SemVer in source will remain at `1.0.0`. This will eventually be changed over to `0.0.0-dev`. Keep in mind that if using the `go get` version of downloading, mockery will not correctly specify its version.

Examples
--------

#### Simplest case

Given this is in `string.go`

```go
package test

type Stringer interface {
	String() string
}
```

Run: `mockery -name=Stringer` and the following will be output to `mocks/Stringer.go`:

```go
package mocks

import "github.com/stretchr/testify/mock"

type Stringer struct {
	mock.Mock
}

func (m *Stringer) String() string {
	ret := m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
```

#### Next level case

See [github.com/jaytaylor/mockery-example](https://github.com/jaytaylor/mockery-example)
for the fully runnable version of the outline below.

```go
package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/jaytaylor/mockery-example/mocks"
	"github.com/stretchr/testify/mock"
)

func main() {
	mockS3 := &mocks.S3API{}

	mockResultFn := func(input *s3.ListObjectsInput) *s3.ListObjectsOutput {
		output := &s3.ListObjectsOutput{}
		output.SetCommonPrefixes([]*s3.CommonPrefix{
			&s3.CommonPrefix{
				Prefix: aws.String("2017-01-01"),
			},
		})
		return output
	}

	// NB: .Return(...) must return the same signature as the method being mocked.
	//     In this case it's (*s3.ListObjectsOutput, error).
	mockS3.On("ListObjects", mock.MatchedBy(func(input *s3.ListObjectsInput) bool {
		return input.Delimiter != nil && *input.Delimiter == "/" && input.Prefix == nil
	})).Return(mockResultFn, nil)

	listingInput := &s3.ListObjectsInput{
		Bucket:    aws.String("foo"),
		Delimiter: aws.String("/"),
	}
	listingOutput, err := mockS3.ListObjects(listingInput)
	if err != nil {
		panic(err)
	}

	for _, x := range listingOutput.CommonPrefixes {
		fmt.Printf("common prefix: %+v\n", *x)
	}
}
```

Imports
-------

mockery pulls in all the same imports used in the file that contains the interface so
that package types will work correctly. It then runs the output through the `imports`
package to remove any unnecessary imports (as they'd result in compile errors).

Types
-----

mockery should handle all types. If you find it does not, please report the issue.

Return Value Provider Functions
--------------------------------

If your tests need access to the arguments to calculate the return values,
set the return value to a function that takes the method's arguments as its own
arguments and returns the return value. For example, given this interface:

```go
package test

type Proxy interface {
  passthrough(ctx context.Context, s string) string
}
```

The argument can be passed through as the return value:

```go
import . "github.com/stretchr/testify/mock"

Mock.On("passthrough", mock.AnythingOfType("context.Context"), mock.AnythingOfType("string")).Return(func(ctx context.Context, s string) string {
    return s
})
```

#### Requirements

`Return` must be passed the same argument count and types as expected by the interface. If the return argument signature of `passthrough` in the above example was instead `(string, error)` in the interface, `Return` would also need a second argument to define the error value.

If any return argument is missing, `github.com/stretchr/testify/mock.Arguments.Get` will emit a panic.

For example, `panic: assert: arguments: Cannot call Get(0) because there are 0 argument(s). [recovered]` indicates that `Return` was not provided any arguments but (at least one) was expected based on the interface. `Get(1)` would indicate that the `Return` call is missing a second argument, and so on.

#### Notes

This approach should be used judiciously, as return values should generally
not depend on arguments in mocks; however, this approach can be helpful for
situations like passthroughs or other test-only calculations.

Name
---------

The `-name` option takes either the name or matching regular expression of interface to generate mock(s) for.

All
---------

It's common for a big package to have a lot of interfaces, so mockery provides `-all`.
This option will tell mockery to scan all files under the directory named by `-dir` ("." by default)
and generates mocks for any interfaces it finds. This option implies `-recursive=true`.

`-all` was designed to be able to be used automatically in the background if required.

Recursive
---------

Use the `-recursive` option to search subdirectories for the interface(s).
This option is only compatible with `-name`. The `-all` option implies `-recursive=true`.

Output
---------

mockery always generates files with the package `mocks` to keep things clean and simple.
You can control which mocks directory is used by using `-output`, which defaults to `./mocks`.

In Package (-inpkg) and KeepTree (-keeptree)
--------------------------------------------

For some complex repositories, there could be multiple interfaces with the same name but in different packages. In that case, `-inpkg` allows generate the mocked interfaces directly in the package that it mocks.

In the case you don't want to generate the mocks into the package but want to keep a similar structure, use the option `-keeptree`.

Specifying File (-filename) and Struct Name (-structname)
---------------------------------------------------------

Use the `-filename` and `-structname` to override the default generated file and struct name.
These options are only compatible with non-regular expressions in -name, where only one mock is generated.

Casing
-------

mockery generates files using the casing of the original interface name.  This
can be modified by specifying `-case=underscore` to format the generated file
name using underscore casing.

Debug
------

Use `mockery -print` to have the resulting code printed out instead of written to disk.

Mocking interfaces in `main`
----------------------------

When your interfaces are in the main package you should supply the `-inpkg` flag.
This will generate mocks in the same package as the target code avoiding import issues.

Semantic Versioning
-------------------

The versioning in this project applies only to the behavior of the mockery binary itself. This project explicitly does not promise a stable internal API, but rather a stable executable. The versioning applies to the following:

1. CLI arguments.
2. Parsing of Golang code. New features in the Golang language will be supported in a backwards-compatible manner, except during major version bumps.
3. Behavior of mock objects. Mock objects can be considered to be part of the public API.
4. Behavior of mockery given a set of argments.

What the version does _not_ track:
1. The interfaces, objects, methods etc. in the vektra/mockery package.
2. Compatibility of `go get`-ing mockery with new or old versions of Golang.
