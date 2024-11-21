Deprecations
=============

`packages`
----------

!!! tip ""

    To resolve this warning, use the [`packages`](features.md#packages-configuration) feature:

    ```yaml title=".mockery.yaml"
    packages:
        [...]
    ```

The [`packages`](features.md#packages-configuration) feature will be the only way to configure mockery in the future.

`issue-845-fix`
---------------

!!! tip ""

    To resolve this warning:
    
    ```yaml title=".mockery.yaml"
    issue-845-fix: True
    ```

This parameter fixes a somewhat uninteresting, but important issue found in [#845](https://github.com/vektra/mockery/issues/845).
In short, mockery ignored the `#!yaml outpkg:` parameter if `#!yaml inpackage:` was set to `#!yaml True`. This prevents users
from being able to set alternate package names for their mocks that are generated in the same directory
as the mocked interface. For example, it's legal Go to append `_test` to the mock package name
if the file is appended with `_test.go` as well. This parameter will be permanently
enabled in mockery v3.

As an example, if you had configuration that looked like this:

```yaml
all: True
dir: "{{.InterfaceDir}}"
mockname: "{{.InterfaceName}}Mock"
outpkg: "{{.PackageName}}_test"
filename: "mock_{{.InterfaceName}}_test.go"
inpackage: True
```

The `#!yaml outpkg` parameter would not be respected and instead would be forced to take on the value of `#!yaml "{{.PackageName}}"`.
To remove the warning, you must set:

```yaml
issue-845-fix: True
```

After this is done, mocks generated in the old scheme will properly respect the `#!yaml outpkg:` parameter previously set
if being generated with `#!yaml inpackage: True`.

`resolve-type-alias`
--------------------

!!! tip ""

    To resolve this warning:
    
    ```yaml title=".mockery.yaml"
    resolve-type-alias: False
    ```

This parameter directs Mockery on whether it should resolve a type alias to its underlying, real
type or if it should generate mocks by referencing. Mockery was changed in [#808](https://github.com/vektra/mockery/pull/808)
to support a new language feature that exposed type aliases in the parsed syntax tree. This meant
that Mockery was now explicitly aware of aliases, which fixed a number of problems:

- [#803](https://github.com/vektra/mockery/pull/803)
- [#331](https://github.com/vektra/mockery/issues/331)

However, it was discovered in [#839](https://github.com/vektra/mockery/issues/839) that this was in fact a backwards-incompatible change. Thus, to maintain backwards compatability guarantees, we created this parameter that will be set to `True` by default.

For all new projects that use Mockery, there is no reason to resolve type aliases so this parameter should almost always
be set to `False`. This will be the permanent behavior in Mockery v3.
