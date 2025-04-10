Templates
=========

Mockery, in its essence, renders templates. This project provides a number of pre-curated templates that you can select with the `#!yaml template:` config parameter. Each template will render a specific kind of implementation of the interface in question. Historically, mocks (used in testing) are common to generate, but anything that implements the interface is fair game.

Mockery provides a few embedded templates you can render, or you can use a URL to specify a remotely-hosted template.

## Template Options

### [`#!yaml template: "testify"`](testify.md#description)

[`testify`](testify.md#description){ data-preview } templates generate powerful, testify-based mock objects. They allow you to create expectations using argument-to-return-value matching logic.

### [`#!yaml template: "matryer"`](matryer.md#description)

[`matryer`](matryer.md#description){ data-preview } templates draw from the mocks generated from the project at https://github.com/matryer/moq. This project was folded into mockery, and thus moq-style mocks can be natively generated from within mockery.

Mocks generated using this template allow you to define precise functions to be run.

### `#!yaml template: "file://"`

You may also provide mockery a path to your own file using the `file://` protocol specifier. The string after `file://` will be the relative or absolute path of your template.

The templates are rendered with the data as shown in the [section below](#template-files).

You can see examples of how the mockery project utilizes the template system to generate the different mock styles:

- [`matryer.templ`](https://github.com/vektra/mockery/blob/v3/internal/mock_matryer.templ)
- [`testify.templ`](https://github.com/vektra/mockery/blob/v3/internal/mock_testify.templ)

### `#!yaml template: "https://"`

You can also host templates remotely. This allows you to specify something like:

```yaml title=""
template: https://raw.githubusercontent.com/vektra/mockery/refs/tags/v3.0.0-beta.8/e2e/test_template_exercise/exercise.templ
template-schema: https://raw.githubusercontent.com/vektra/mockery/refs/tags/v3.0.0-beta.8/e2e/test_template_exercise/exercise.templ.schema.json
```

!!! note "Note on Versioning"

    Templates hosted remotely should be versioned just like any other piece of code. We recommend hosting the template on a git repo so that you may reference explicit tags.

    It's best practice to maintain 3 sets of tags:

    1. Major: `v3`
    2. Minor: `v3.1`
    3. Patch: `v3.1.2`

    This will give consumers of your template the ability to specify which level of granularity they want to track versions.

The non-TLS variant is also supported: `#!yaml template: "http://"`, but not recommended for obvious reasons.

## Schemas

Templates can provide a JSON Schema file that describes the format of the `TemplateData` parameter. Mockery auto-discovers the location of these schema files by appending `.schema.json` to the path of the template. For example, if you provide to mockery `#!yaml template: file://./path/to/template.tmpl`, it will look for a file at `file://./path/to/template.tmpl.schema.json`. If found, this schema will be applied to the `TemplateData` type sent to the template.

To get started with JSON Schema, you can borrow an example JSON document used for the mockery project itself:

```json title="schema.json"
{
    "$schema": "http://json-schema.org/draft-07/schema#",
    "title": "vektra/mockery testify mock",
    "type": "object",
    "additionalProperties": false,
    "properties": {
      "boilerplate-file": {
        "type": "string"
      },
      "mock-build-tags": {
        "type": "string"
      },
      "unroll-variadic": {
        "type": "boolean"
      }
    },
    "required": []
}
```

Note that the `#!json "additionalProperties": false` parameter is crucial to ensure only the specified parameters exist in the configured `#!yaml template-data: {}` map.

!!! tip "`template-schema`"

    You can specify a custom schema path using the [`#!yaml template-schema:`](../configuration.md#parameter-descriptions)parameter.

## Template Data

Templates are rendered with functions and data you can utilize to generate your mocks. Links are shown below:

| Description | Link |
|-|-|
| Functions | [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/vektra/mockery/v3/template_funcs#pkg-variables) | 
| Data | [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/vektra/mockery/v3/template#Data) |
