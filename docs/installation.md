Getting Started
================

Installation
-------------

### go install

Supported, but not recommended: [see wiki page](https://github.com/vektra/mockery/wiki/Installation-Methods#go-install) and [related discussions](https://github.com/vektra/mockery/pull/456).

Alternatively, you can use the go install method to compile the project using your local environment:

    go install github.com/vektra/mockery/v2@latest

### GitHub Release

<small>recommended</small>

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

