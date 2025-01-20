Templates
=========

Mockery, in its essence, renders templates. This project provides a number of pre-curated
templates that you can select with the `#!yaml template:` config parameter.

## Template Options

### `#!yaml template: "testify"`

[`testify`](template-testify.md) templates generate powerful, testify-based mock objects. They allow you to create expectations using argument-to-return-value matching logic.

```go
package test

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestRequesterMock(t *testing.T) {
    m := NewMockRequester(t)
    m.EXPECT().Get("foo").Return("bar", nil).Once()
    retString, err := m.Get("foo")
    assert.NoError(t, err)
    assert.Equal(t, retString, "bar")
}
```

### `#!yaml template: "matryer"`

[`matryer`](template-matryer.md) templates draw from the mocks generated from the project at https://github.com/matryer/moq. This project was folded into mockery, and thus moq-style mocks can be natively generated from within mockery.

Mocks generated using this template allow you to define precise functions to be run. Example:

```go
func TestRequesterMoq(t *testing.T) {
    m := &MoqRequester{
        GetFunc: func(path string) (string, error) {
            fmt.Printf("Go path: %s\n", path)
            return path + "/foo", nil
        },
    }
    result, err := m.Get("/path")
    assert.NoError(t, err)
    assert.Equal(t, "/path/foo", result)
}
```

### `#!yaml template: "file://`

You may also provide mockery a path to your own file using the `file://` protocol specifier. The string after `file://` will be the relative or absolute path of your template.

The templates are rendered with the data as shown in the [section below](#template-files).

You can see examples of how the mockery project utilizes the template system to generate the different mock styles:

- [`moq.templ`](https://github.com/vektra/mockery/blob/v3/internal/moq.templ)
- [`mockery.templ`](https://github.com/vektra/mockery/blob/v3/internal/mockery.templ)

## Data Provided To Templates

Mockery has two separate template instances: one for the `.mockery.yml` file, and one for the mock templates. Each instance has a different set of variables and functions available to it. All functions are [pipeline-compatible](https://pkg.go.dev/text/template#hdr-Pipelines).

### `.mockery.yml`

As seen in the [configuration](configuration.md) section, mockery configuration has template variables and methods available to it.

#### Functions

Functions provided are in the [`StringManipulationFuncs`](https://pkg.go.dev/github.com/vektra/mockery/v3/template#pkg-variables) variable.

#### Variables

The variables available are defined in the [`template.ConfigData`](https://pkg.go.dev/github.com/vektra/mockery/v3/template#ConfigData) struct.


### Template Files

#### Functions

Template files have both [`StringManipulationFuncs`](https://pkg.go.dev/github.com/vektra/mockery/v3/template#pkg-variables) and [`TemplateMockFuncs`](https://pkg.go.dev/github.com/vektra/mockery/v3@v3.0.0-alpha.10/template#pkg-variables) available as functions.

#### Variables

The template is supplied with the [`template.Data`](https://pkg.go.dev/github.com/vektra/mockery/v3/template#Data) struct. Some attributes return types such as [`template.MockData`](https://pkg.go.dev/github.com/vektra/mockery/v3@v3.0.0-alpha.10/template#MockData) and [`template.Package`](https://pkg.go.dev/github.com/vektra/mockery/v3/template#Package) which themselves contain methods that may also be called.
