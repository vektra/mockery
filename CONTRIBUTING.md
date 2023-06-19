# Contributing

Thank you for investing your time in contributing to our project!

Read our [Code of Conduct](https://github.com/vektra/mockery/blob/master/CODE_OF_CONDUCT.md) to keep our community approachable and respectable.

## Local development setup

All of the local development tools are go-based and are versioned in our go.mod file. Simply call `go download -x` to initialize and download all of our tooling.

This project uses Taskfile, a better alternative to Makefile. Run `task -l` for list of valid targets.

## Working with documentation

We use [mkdocs](https://www.mkdocs.org/) with the [mkdocs-material theme](https://squidfunk.github.io/mkdocs-material/).

To preview the documentation locally, run `task mkdocs.serve`. The task will install the required mkdocs plugins and theme
and run the mkdocs server with real-time updating/refreshing.

## Submitting PRs

Before submitting PRs, it's strongly recommended you create an issue to discuss the problem, and possible solutions. PRs added without any discussion run the risk of not being accepted, which is a waste of your time. Issues marked with `approved feature` are features that the maintainers are willing to accept into the code, and are approved for development. PRs that implement an `approved feature` have a high likelihood of being accepted.
