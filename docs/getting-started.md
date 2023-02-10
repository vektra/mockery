Getting Started
================

Installation
-------------

### go install

Supported, but not recommended: [see wiki page](https://github.com/vektra/mockery/wiki/Installation-Methods#go-install) and [related discussions](https://github.com/vektra/mockery/pull/456).

Alternatively, you can use the go install method to compile the project using your local environment:

    go install github.com/vektra/mockery/v2@latest

### GitHub Release

Visit the [releases page](https://github.com/vektra/mockery/releases) to download one of the pre-built binaries for your platform.

### Docker

Use the [Docker image](https://hub.docker.com/r/vektra/mockery)

    docker pull vektra/mockery

Generate all the mocks for your project:

	docker run -v "$PWD":/src -w /src vektra/mockery --all

### Homebrew

Install through [brew](https://brew.sh/)

    brew install mockery
    brew upgrade mockery

Configuration
--------------

mockery uses [spf13/viper](https://github.com/spf13/viper) under the hood for its configuration parsing. It is bound to three different configuration sources, in order of decreasing precedence:

1. Command line
2. Environment variables
3. Configuration file

Copy the recommended basic configuration to a file called `.mockery.yaml` at the top-level of your repo:

```yaml
all: True
keeptree: True
```

mockery will search upwards from your current-working-directory up to the root path, so the same configuration should be able to follow you within your project.

Run mockery
------------

### For all interfaces in project

```bash
$ mockery
09 Feb 23 22:47 CST INF Starting mockery dry-run=false version=v2.18.0
09 Feb 23 22:47 CST INF Using config: /Users/landonclipp/git/LandonTClipp/mockery/.mockery.yaml dry-run=false version=v2.18.0
09 Feb 23 22:47 CST INF Walking dry-run=false version=v2.18.0
09 Feb 23 22:47 CST INF Generating mock dry-run=false interface=A qualified-name=github.com/vektra/mockery/v2/pkg/fixtures version=v2.18.0
```

### Using `go generate`

`go generate` is often preferred as it give you more targeted generation of specific interfaces. Use `generate` as a directive above the interface you want to generate a mock for.

``` golang
package example_project

//go:generate mockery --name Root --all=False
type Root interface {
        Foobar(s string) error
}
```

Then simply:

``` bash
$ go generate      
09 Feb 23 22:55 CST INF Starting mockery dry-run=false version=v2.18.0
09 Feb 23 22:55 CST INF Using config: /Users/landonclipp/git/LandonTClipp/mockery/.mockery.yaml dry-run=false version=v2.18.0
09 Feb 23 22:55 CST INF Walking dry-run=false version=v2.18.0
09 Feb 23 22:55 CST INF Generating mock dry-run=false interface=Root qualified-name=github.com/vektra/mockery/v2/pkg/fixtures/example_project version=v2.18.0
```

Note that mockery running in `go generate` will still ingest configuration from your top-level `.mockery.yaml` file, so you may have to enable/disable certain configuration parameters from the command line to prevent collisions.
