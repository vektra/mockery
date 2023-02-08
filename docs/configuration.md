Configuration
==============

mockery uses [spf13/viper](https://github.com/spf13/viper) under the hood for its configuration parsing. It is bound to three different configuration sources, in order of decreasing precedence:

1. Command line
2. Environment variables
3. Configuration file

If a parameter is named `with-expecter` and we want a value of `True`, then these are the formats for each source:

| source | value |
|--------|-------|
| command line | `--with-expecter=true` |
| Environment variable | `MOCKERY_WITH_EXPECTER=True` |
| yaml | `with-expecter: True` |

Recommended Basic Config
-------------------------

Copy the recommended basic configuration to a file called `.mockery.yaml` at the top-level of your repo:

```yaml title=".mockery.yaml"
inpackage: True
testonly: True
with-expecter: True
keeptree: False
```

mockery will search upwards from your current-working-directory up to the root path, so the same configuration should be able to follow you within your project.

Parameter Descriptions
-----------------------

| name | description |
|------|-------------|
| `name`  | The `name` option takes either the name or matching regular expression of the interface to generate mock(s) for. |
| `all`  |  It's common for a big package to have a lot of interfaces, so mockery provides `all`. This option will tell mockery to scan all files under the directory named by `--dir` ("." by default) and generates mocks for any interfaces it finds. This option implies `recursive: True`. |
| `recursive`  |  Use the `recursive` option to search subdirectories for the interface(s). This option is only compatible with `name`. The `all` option implies `recursive: True`. |
| `output` | mockery always generates files with the package `mocks` to keep things clean and simple. You can control which mocks directory is used by using `output`, which defaults to `./mocks`. |
|`outpkg`| Use `outpkg` to specify the package name of the generated mocks.|
| `inpackage` and `keeptree` | For some complex repositories, there could be multiple interfaces with the same name but in different packages. In that case, `inpackage` allows generating the mocked interfaces directly in the package that it mocks. In the case you don't want to generate the mocks into the package but want to keep a similar structure, use the option `keeptree`. |
| `filename` | Use the `filename` and `structname` to override the default generated file and struct name. These options are only compatible with non-regular expressions in `name`, where only one mock is generated. |
| `case` | mockery generates files using the casing of the original interface name.  This can be modified by specifying `case: underscore` to format the generated file name using underscore casing. |
| `print` | Use `print: True` to have the resulting code printed out instead of written to disk. |
| `exported` | Use `exported: True` to generate public mocks for private interfaces. |
| `with-expecter` | Use `with-expecter: True` to generate `EXPECT()` methods for your mocks. This is the preferred way to setup your mocks. |
| `testonly` | Prepend every mock file with `_test.go`. This is useful in cases where you are generating mocks `inpackage` but don't want the mocks to be visible to code outside of tests. |
| `inpackage-suffix` | When `inpackage-suffix` is set to `True`, mock files are suffixed with `_mock` instead of being prefixed with `mock_` for InPackage mocks |
| `replace-type source=destination` | Replaces aliases, packages and/or types during generation.|