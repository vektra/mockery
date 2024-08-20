Configuration
==============

mockery uses [spf13/viper](https://github.com/spf13/viper) under the hood for its configuration parsing. 

Merging Precedence
------------------

The configuration applied to a specific mocked interface is merged according to the following precedence (in decreasing priority):

1. Interface-specific config in `.mockery.yaml`
2. Package-specific config in `.mockery.yaml`
3. Command-line options
4. Environment variables
5. Top-level defaults in `.mockery.yaml`

Formatting
----------

If a parameter is named `with-expecter` and we want a value of `True`, then these are the formats for each source:

| source               | value                        |
|----------------------|------------------------------|
| command line         | `--with-expecter=true`       |
| Environment variable | `MOCKERY_WITH_EXPECTER=True` |
| yaml                 | `with-expecter: True`        |

Recommended Basic Config
-------------------------

Copy the recommended basic configuration to a file called `.mockery.yaml` at the top-level of your repo:

```yaml title=".mockery.yaml"
with-expecter: true
packages:
    github.com/your-org/your-go-project:
        # place your package-specific config here
        config:
        interfaces:
            # select the interfaces you want mocked
            Foo:
                # Modify package-level config for this specific interface (if applicable)
                config:
```

mockery will search upwards from your current-working-directory up to the root path, so the same configuration should be able to follow you within your project.

See the [`features` section](features.md#examples) for more details on how the config is structured.

Parameter Descriptions
-----------------------

!!! success "new style `packages` config"
    The [`packages`]() config section is the new style of configuration. All old config semantics, including `go:generate` and any config files lacking the `packages` section is officially deprecated as of v2.31.0. Legacy semantics will be completely removed in v3.

    Please see the [features section](features.md#packages-configuration) for more details on how `packages` works, including some example configuration.

    Please see the [migration docs](migrating_to_packages.md) for details on how to migrate your config.

| name                                                   | templated                 | default                               | description                                                                                                                                                                                                                                          |
|--------------------------------------------------------|---------------------------|---------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `all`                                                  | :fontawesome-solid-x:     | `#!yaml false`                        | Generate all interfaces for the specified packages.                                                                                                                                                                                                  |
| `boilerplate-file`                                     | :fontawesome-solid-x:     | `#!yaml ""`                           | Specify a path to a file that contains comments you want displayed at the top of all generated mock files. This is commonly used to display license headers at the top of your source code.                                                          |
| `config`                                               | :fontawesome-solid-x:     | `#!yaml ""`                           | Set the location of the mockery config file.                                                                                                                                                                                                         |
| `dir`                                                  | :fontawesome-solid-check: | `#!yaml "mocks/{{.PackagePath}}"`     | The directory where the mock file will be outputted to.                                                                                                                                                                                              |
| `disable-config-search`                                | :fontawesome-solid-x:     | `#!yaml false`                        | Disable searching for configuration files                                                                                                                                                                                                            |
| `disable-func-mocks`                                   | :fontawesome-solid-x:     | `#!yaml false`                        | Disable generation of function mocks.                                                                                                                                                                                                                |
| `disable-version-string`                               | :fontawesome-solid-x:     | `#!yaml false`                        | Disable the version string in the generated mock files.                                                                                                                                                                                              |
| `dry-run`                                              | :fontawesome-solid-x:     | `#!yaml false`                        | Print the actions that would be taken, but don't perform the actions.                                                                                                                                                                                |
| `exclude`                                              | :fontawesome-solid-x:     | `#!yaml []`                           | Specify subpackages to exclude when using `#!yaml recursive: True`                                                                                                                                                                                   |
| `exclude-regex`                                        | :fontawesome-solid-x:     | `#!yaml ""`                           | When set along with `include-regex`, then interfaces which match `include-regex` but also match `exclude-regex` will not be generated. If `all` is set, or if `include-regex` is not set, then `exclude-regex` has no effect.                        |
| `filename`                                             | :fontawesome-solid-check: | `#!yaml "mock_{{.InterfaceName}}.go"` | The name of the file the mock will reside in.                                                                                                                                                                                                        |
| `include-auto-generated`                               | :fontawesome-solid-x:     | `#!yaml true`                         | Set to `#!yaml false` if you need mockery to skip auto-generated files during its recursive package discovery. When set to `#!yaml true`, mockery includes auto-generated files when determining if a particular directory is an importable package. |
| `include-regex`                                        | :fontawesome-solid-x:     | `#!yaml ""`                           | When set, only interface names that match the expression will be generated. This setting is ignored if `all: True` is specified in the configuration. To further refine the interfaces generated, use `exclude-regex`.                               |
| `inpackage`                                            | :fontawesome-solid-x:     | `#!yaml false`                        | When generating mocks alongside the original interfaces, you must specify `inpackage: True` to inform mockery that the mock is being placed in the same package as the original interface.                                                           |
| `log-level`                                            | :fontawesome-solid-x:     | `#!yaml "info"`                       | Set the level of the logger                                                                                                                                                                                                                          |
| `mock-build-tags`                                      | :fontawesome-solid-x:     | `#!yaml ""`                           | Set the build tags of the generated mocks. Read more about the [format](https://pkg.go.dev/cmd/go#hdr-Build_constraints).                                                                                                                             |
| `mockname`                                             | :fontawesome-solid-check: | `#!yaml "Mock{{.InterfaceName}}"`     | The name of the generated mock.                                                                                                                                                                                                                      |
| `outpkg`                                               | :fontawesome-solid-check: | `#!yaml "{{.PackageName}}"`           | Use `outpkg` to specify the package name of the generated mocks.                                                                                                                                                                                     |
| [`packages`](features.md#packages-configuration)       | :fontawesome-solid-x:     | `#!yaml null`                         | A dictionary containing configuration describing the packages and interfaces to generate mocks for.                                                                                                                                                  |
| `print`                                                | :fontawesome-solid-x:     | `#!yaml false`                        | Use `print: True` to have the resulting code printed out instead of written to disk.                                                                                                                                                                 |
| [`recursive`](features.md#recursive-package-discovery) | :fontawesome-solid-x:     | `#!yaml false`                        | When set to `true` on a particular package, mockery will recursively search for all sub-packages and inject those packages into the config map.                                                                                                      |
| [`replace-type`](features.md#replace-types)            | :fontawesome-solid-x:     | `#!yaml null`                         | Replaces aliases, packages and/or types during generation.                                                                                                                                                                                           |
| `tags`                                                 | :fontawesome-solid-x:     | `#!yaml ""`                           | A space-separated list of additional build tags to load packages.                                                                                                                                                                                                                |
| [`with-expecter`](features.md#expecter-structs)        | :fontawesome-solid-x:     | `#!yaml true`                         | Use `with-expecter: True` to generate `EXPECT()` methods for your mocks. This is the preferred way to set up your mocks.                                                                                                                             |

Layouts
-------

Using different configuration parameters, we can deploy our mocks on-disk in various ways. These are some common layouts:

!!! info "layouts"

    === "defaults"

        ```yaml
        filename: "mock_{{.InterfaceName}}.go"
        dir: "mocks/{{.PackagePath}}"
        mockname: "Mock{{.InterfaceName}}"
        outpkg: "{{.PackageName}}"
        ```

        If these variables aren't specified, the above values will be applied to the config options. This strategy places your mocks into a separate `mocks/` directory.

        **Interface Description**

        | name | value |
        |------|-------|
        | `InterfaceName` | `MyDatabase` |
        | `PackagePath` | `github.com/user/project/pkgName` |
        | `PackageName` | `pkgName` |

        **Output**

        The mock will be generated at:

        ```
        mocks/github.com/user/project/pkgName/mock_MyDatabase.go
        ```

        The mock file will look like:

        ```go
        package pkgName

        import mock "github.com/stretchr/testify/mock"

        type MockMyDatabase struct {
          mock.Mock
        }
        ```
    === "adjacent to interface"
    
    	!!! warning

            Mockery does not protect against modifying original source code. Do not generate mocks using this config with uncommitted code changes.


        ```yaml
        filename: "mock_{{.InterfaceName}}.go"
        dir: "{{.InterfaceDir}}"
        mockname: "Mock{{.InterfaceName}}"
        outpkg: "{{.PackageName}}"
        inpackage: True
        ```

        Instead of the mocks being generated in a different folder, you may elect to generate the mocks alongside the original interface in your package. This may be the way most people define their configs, as it removes circular import issues that can happen with the default config.

        For example, the mock might be generated along side the original source file like this:

        ```
        ./path/to/pkg/db.go
        ./path/to/pkg/mock_MyDatabase.go
        ```

        **Interface Description**

        | name | value |
        |------|-------|
        | `InterfaceName` | `MyDatabase` |
        | `PackagePath` | `github.com/user/project/path/to/pkg`
        | `PackagePathRelative` | `path/to/pkg` |
        | `PackageName` | `pkgName` |
        | `SourceFile` | `./path/to/pkg/db.go` |

        **Output**

        Mock file will be generated at:

        ```
        ./path/to/pkg/mock_MyDatabase.go
        ```

        The mock file will look like:

        ```go
        package pkgName

        import mock "github.com/stretchr/testify/mock"

        type MockMyDatabase struct {
          mock.Mock
        }
        ```

Templated Strings
------------------

mockery configuration makes use of the Go templating system. 

### Variables

!!! note
    Templated variables are only available when using the `packages` config feature.

Variables that are marked as being templated are capable of using mockery-provided template parameters.

| name                    | description                                                                                                                                                                                                                                                                                   |
|-------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| InterfaceDir            | The directory path of the original interface being mocked. This can be used as <br>`#!yaml dir: "{{.InterfaceDir}}"` to place your mocks adjacent to the original interface. This should not be used for external interfaces.                                                                 |
| InterfaceDirRelative    | The directory path of the original interface being mocked, relative to the current working directory. If the path cannot be made relative to the current working directory, this variable will be set equal to `PackagePath`                                                                  |
| InterfaceFile           | The file path of the original interface being mocked. **NOTE:** This option will only write one mock implementation to the output file. If multiple mocks are defined in your original file, only one mock will be written to the output.                                                     |
| InterfaceName           | The name of the original interface being mocked                                                                                                                                                                                                                                               |
| InterfaceNameCamel      | Converts a string `interface_name` to `InterfaceName`. <br /><b style="color:var(--md-code-hl-number-color);">DEPRECATED</b>: use `{{ .InterfaceName | camelcase }}` instead                                                                                                                                                                     | 
| InterfaceNameLowerCamel | Converts `InterfaceName` to `interfaceName` . <br /><b style="color:var(--md-code-hl-number-color);">DEPRECATED</b>: use `{{ .InterfaceName | camelcase | firstLower }}` instead                                                                                                                                                                |
| InterfaceNameSnake      | Converts `InterfaceName` to `interface_name` . <br /><b style="color:var(--md-code-hl-number-color);">DEPRECATED</b>: use `{{ .InterfaceName | snakecase }}` instead                                                                                                                                                                             |
| InterfaceNameLower      | Converts `InterfaceName` to `interfacename` . <br /><b style="color:var(--md-code-hl-number-color);">DEPRECATED</b>: use `{{ .InterfaceName | lower }}` instead                                                                                                                                                                                  |
| Mock                    | A string that is `Mock` if the interface is exported, or `mock` if it is not exported. Useful when setting the name of your mock to something like: <br>`#!yaml mockname: "{{.Mock}}{{.InterfaceName}}"`<br> This way, the mock name will retain the exported-ness of the original interface. |
| MockName                | The name of the mock that will be generated. Note that this is simply the `mockname` configuration variable                                                                                                                                                                                   |
| PackageName             | The name of the package from the original interface                                                                                                                                                                                                                                           |
| PackagePath             | The fully qualified package path of the original interface                                                                                                                                                                                                                                    |

### Functions

!!! note
    Templated functions are only available when using the `packages` config feature.

Template functions allow you to inspect and manipulate template variables.

All template functions are calling native Go functions under the hood, so signatures and return values matches the Go functions you are probably already familiar with.

To learn more about the templating syntax, please [see the Go `text/template` documentation](https://pkg.go.dev/text/template)

* [`contains` string substr](https://pkg.go.dev/strings#Contains)
* [`hasPrefix` string prefix](https://pkg.go.dev/strings#HasPrefix)
* [`hasSuffix` string suffix](https://pkg.go.dev/strings#HasSuffix)
* [`join` elems sep](https://pkg.go.dev/strings#Join)
* [`replace` string old new n](https://pkg.go.dev/strings#Replace)
* [`replaceAll` string old new](https://pkg.go.dev/strings#ReplaceAll)
* [`split` string sep](https://pkg.go.dev/strings#Split)
* [`splitAfter` string sep](https://pkg.go.dev/strings#SplitAfter)
* [`splitAfterN` string sep n](https://pkg.go.dev/strings#SplitAfterN)
* [`trim` string cutset](https://pkg.go.dev/strings#Trim)
* [`trimLeft` string cutset](https://pkg.go.dev/strings#TrimLeft)
* [`trimPrefix` string prefix](https://pkg.go.dev/strings#TrimPrefix)
* [`trimRight` string cutset](https://pkg.go.dev/strings#TrimRight)
* [`trimSpace` string](https://pkg.go.dev/strings#TrimSpace)
* [`trimSuffix` string suffix](https://pkg.go.dev/strings#TrimSuffix)
* [`lower` string](https://pkg.go.dev/strings#ToLower)
* [`upper` string](https://pkg.go.dev/strings#ToUpper)
* [`camelcase` string](https://pkg.go.dev/github.com/huandu/xstrings#ToCamelCase)
* [`snakecase` string](https://godoc.org/github.com/huandu/xstrings#ToSnakeCase)
* [`kebabcase` string](https://godoc.org/github.com/huandu/xstrings#ToKebabCase)
* [`firstLower` string](https://godoc.org/github.com/huandu/xstrings#FirstRuneToLower)
* [`firstUpper` string](https://godoc.org/github.com/huandu/xstrings#FirstRuneToUpper)
* [`matchString` pattern](https://pkg.go.dev/regexp#MatchString)
* [`quoteMeta` string](https://pkg.go.dev/regexp#QuoteMeta)
* [`base` string](https://pkg.go.dev/path/filepath#Base)
* [`clean` string](https://pkg.go.dev/path/filepath#Clean)
* [`dir` string](https://pkg.go.dev/path/filepath#Dir)
* [`expandEnv` string](https://pkg.go.dev/os#ExpandEnv)
* [`getenv` string](https://pkg.go.dev/os#Getenv)

Legacy config options
---------------------

??? danger "legacy configuration options"

    The legacy config options will be removed in v3 and are deprecated (but supported) in v2.
    
    | name | description |
    |------|-------------|
    | `all`  |  It's common for a big package to have a lot of interfaces, so mockery provides `all`. This option will tell mockery to scan all files under the directory named by `--dir` ("." by default) and generates mocks for any interfaces it finds. This option implies `recursive: True`. |
    | `boilerplate-file` | Specify a path to a file that contains comments you want displayed at the top of all generated mock files. This is commonly used to display license headers at the top of your source code. |
    | `case` | mockery generates files using the casing of the original interface name.  This can be modified by specifying `case: underscore` to format the generated file name using underscore casing. |
    | `exclude` | This parameter is a list of strings representing path prefixes that should be excluded from mock generation. |
    | `exported` | Use `exported: True` to generate public mocks for private interfaces. |
    | `filename` | Use the `filename` and `structname` to override the default generated file and struct name. These options are only compatible with non-regular expressions in `name`, where only one mock is generated. |
    | `inpackage-suffix` | When `inpackage-suffix` is set to `True`, mock files are suffixed with `_mock` instead of being prefixed with `mock_` for InPackage mocks |
    | `inpackage` and `keeptree` | For some complex repositories, there could be multiple interfaces with the same name but in different packages. In that case, `inpackage` allows generating the mocked interfaces directly in the package that it mocks. In the case you don't want to generate the mocks into the package but want to keep a similar structure, use the option `keeptree`. |
    | `name`  | The `name` option takes either the name or matching regular expression of the interface to generate mock(s) for. |
    | `output` | mockery always generates files with the package `mocks` to keep things clean and simple. You can control which mocks directory is used by using `output`, which defaults to `./mocks`. |
    |`outpkg`| Use `outpkg` to specify the package name of the generated mocks.|
    | `print` | Use `print: True` to have the resulting code printed out instead of written to disk. |
    | `recursive`  |  Use the `recursive` option to search subdirectories for the interface(s). This option is only compatible with `name`. The `all` option implies `recursive: True`. |
    | `replace-type source=destination` | Replaces aliases, packages and/or types during generation.|
    | `testonly` | Prepend every mock file with `_test.go`. This is useful in cases where you are generating mocks `inpackage` but don't want the mocks to be visible to code outside of tests. |
    | `with-expecter` | Use `with-expecter: True` to generate `EXPECT()` methods for your mocks. This is the prefervar(--md-code-hl-number-color) way to setup your mocks. |
