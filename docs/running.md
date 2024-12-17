Running
========

If your `.mockery.yaml` file has been populated with the packages and interfaces you want mocked, mockery can be run with no arguments. Take for example how the mockery project itself is configured:

```yaml
quiet: False
keeptree: True
disable-version-string: True
with-expecter: True
mockname: "{{.InterfaceName}}"
filename: "{{.MockName}}.go"
outpkg: mocks
packages:
  github.com/vektra/mockery/v2/pkg:
    interfaces:
      TypesPackage:
# Lots more config...
```

From anywhere within your repo, you can simply call `mockery` once, and it will find your config either by respecting the `#!yaml config` path you gave it, or by searching upwards from the current working directory.

```bash
mockery
08 Jul 23 01:40 EDT INF Starting mockery dry-run=false version=v2.31.0
08 Jul 23 01:40 EDT INF Using config: /Users/landonclipp/git/LandonTClipp/mockery/.mockery.yaml dry-run=false version=v2.31.0
```

!!! question "Command line arguments"
    It is valid to specify arguments from the command line. The configuration precedence is specified in the [Configuration](configuration.md#merging-precedence) docs.
