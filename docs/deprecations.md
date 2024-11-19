Deprecations
=============

`packages`
----------

The [`packages`](features.md#packages-configuration) feature will be the only way to configure mockery in the future.

`issue-845-fix`
---------------

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
