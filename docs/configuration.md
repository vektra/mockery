Configuration
==============

Example
-------

All configuration is specified in a `.mockery.yml` file. An example config file may look like this:

```yaml
all: False
template-data:
  boilerplate-file: ./path/to/boilerplate.txt
template: mockery
packages:
  github.com/vektra/example:
    config:
      # Make use of the template variables to place the mock in the same
      # directory as the original interface.
      dir: "{{.InterfaceDir}}"
      filename: "mocks_test.go"
      outpkg: "{{.PackageName}}_test"
      mockname: "Mock{{.InterfaceName}}"
    interfaces:
      Foo:
      Bar:
        config:
          # Make it unexported instead
          mockname: "mock{{.InterfaceName}}"
      Baz:
        # Create two mock implementations of Baz with different names.
        configs:
          - filename: "mocks_baz_one_test.go"
            mockname: "MockBazOne"
          - filename: "mocks_baz_two_test.go"
            mockname: "MockBazTwo"
  io:
    config:
      dir: path/to/io/mocks
      filename: "mocks_io.go"

```

These are the highlights of the config scheme:

1. The parameters are merged hierarchically
2. There are a number of template variables available to generalize config values.
3. The style of mock to be generated is specified using the [`template`](templates/index.md) parameter.

An output file may contain multiple mocks, but the only rule is that all the mocks in the file must come from the same package. Because of this, mocks for different packages must go in different files.

Parameter Descriptions
-----------------------

| name                                                   | templated                 | default                               | description                                                                                                                                                                                                                                          |
|--------------------------------------------------------|---------------------------|---------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `all`                                                  | :fontawesome-solid-x:     | `#!yaml false`                        | Generate all interfaces for the specified packages.                                                                                                                                                                                                  |
| `_anchors`                                             | :fontawesome-solid-x:     | `#!yaml {}`                           | Unused by mockery, but allowed in the config schema so that you may define arbitrary yaml anchors.                                                                                                                                                   |
| `config`                                               | :fontawesome-solid-x:     | `#!yaml ""`                           | Set the location of the mockery config file.                                                                                                                                                                                                         |
| `dir`                                                  | :fontawesome-solid-check: | `#!yaml "mocks/{{.SrcPackagePath}}"`  | The directory where the mock file will be outputted to.                                                                                                                                                                                              |
| `exclude-subpkg-regex`                                 | :fontawesome-solid-x:     | `#!yaml []`                           | A list of regular expressions that denote which subpackages should be excluded when `#!yaml recursive: true` |
| `exclude-regex`                                        | :fontawesome-solid-x:     | `#!yaml ""`                           | When set along with `include-regex`, then interfaces which match `include-regex` but also match `exclude-regex` will not be generated. If `all` is set, or if `include-regex` is not set, then `exclude-regex` has no effect.                        |
| `filename`                                             | :fontawesome-solid-check: | `#!yaml "mock_{{.InterfaceName}}.go"` | The name of the file the mock will reside in.                                                                                                                                                                                                        |
| `force-file-write`                                     | :fontawesome-solid-x:     | `#!yaml false`                        | When set to `#!yaml force-file-write: true`, mockery will forcibly overwrite any existing files. |
| `formatter`                                            | :fontawesome-solid-x:     | `#!yaml "goimports"`                  | The formatter to use on the rendered template. Choices are: `gofmt`, `goimports`, `noop`.                                                                                                                                                            |
| `include-regex`                                        | :fontawesome-solid-x:     | `#!yaml ""`                           | When set, only interface names that match the expression will be generated. This setting is ignored if `all: True` is specified in the configuration. To further refine the interfaces generated, use `exclude-regex`.                               |
| `log-level`                                            | :fontawesome-solid-x:     | `#!yaml "info"`                       | Set the level of the logger                                                                                                                                                                                                                          |
| `mockname`                                             | :fontawesome-solid-check: | `#!yaml "Mock{{.InterfaceName}}"`     | The name of the generated mock.                                                                                                                                                                                                                      |
| `outpkg`                                               | :fontawesome-solid-check: | `#!yaml "{{.PackageName}}"`           | Use `outpkg` to specify the package name of the generated mocks.                                                                                                                                                                                     |
| [`packages`](features.md#packages-configuration)       | :fontawesome-solid-x:     | `#!yaml null`                         | A dictionary containing configuration describing the packages and interfaces to generate mocks for.                                                                                                                                                  |
| `pkgname`                                              | :fontawesome-solid-check: | `#!yaml "{{.SrcPackageName}}"         | The `#!go package name` given to the generated mock files.                                                                                                                                                                                           |
| [`recursive`](features.md#recursive-package-discovery) | :fontawesome-solid-x:     | `#!yaml false`                        | When set to `true` on a particular package, mockery will recursively search for all sub-packages and inject those packages into the config map.                                                                                                      |
| [`replace-type`](replace-type.md)                      | :fontawesome-solid-x:     | `#!yaml {}`                           | Use this parameter to specify type replacements.                                 |
| `tags`                                                 | :fontawesome-solid-x:     | `#!yaml ""`                           | A space-separated list of additional build tags to load packages.                                                                                                                                                                                    |
| `template`                                             | :fontawesome-solid-x:     | `#!yaml ""`                           | The template to use. The choices are `moq`, `mockery`, or a file path provided by `file://path/to/file.txt`.                                                                                                                                         |
| `template-data`                                        | :fontawesome-solid-x:     | `#!yaml {}`                           | A `map[string]any` that provides arbitrary options to the template. Each template will have a different set of accepted keys. Refer to each template's documentation for more details.                                                               |


Templates
---------

Parameters marked as being templated have access to a number of template variables and functions through the Go [`text/template`](https://pkg.go.dev/text/template#hdr-Examples) system.

### Variables

The variables provided are specified in the [`config.Data`](https://pkg.go.dev/github.com/vektra/mockery/v3/config#Data) struct.

### Functions

All of the functions defined in [`StringManipulationFuncs`](https://pkg.go.dev/github.com/vektra/mockery/v3/shared#pkg-variables) are available to templated parameters.

Merging Precedence
------------------

The configuration applied to a specific mocked interface is merged according to the following precedence (in increasing priority):

1. Top-level defaults in `.mockery.yaml`
2. Environment variables
3. Command-line options
4. Package-specific config in `.mockery.yaml`
5. Interface-specific config in `.mockery.yaml`

Formatting
----------

If a parameter is named `enable-feature` and we want a value of `True`, then these are the formats for each source:

| source               | value                        |
|----------------------|------------------------------|
| command line         | `--enable-feature=true`       |
| Environment variable | `MOCKERY_ENABLE_FEATURE=True` |
| yaml                 | `#!yaml enable-feature: True` |

