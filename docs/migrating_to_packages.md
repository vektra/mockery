---
title: Migrating To Packages
---

The [packages](/mockery/features/#packages-configuration) feature is a new configuration scheme that aims to simplify and improve a lot of legacy behavior. This will be the only way to generate mocks in v3. These docs outline general principals for migrating to the new scheme.

Background
----------

mockery was built during the pre-module era of Golang. Much of its codebase and configuration syntax was designed around file-based operations. This model became highly inefficient once Golang migrated to module-based packages. The old configuration semantics also proved limiting -- many users introduced and requested feature additions to mockery to support esoteric use-cases. This proved to be a huge maintenance burden that existed solely because the configuration model could not flexibly describe all the situations users wanted. The `packages` semantics provides us a few highly desirable traits:

1. Orders of magnitude performance increase, due to calling `packages.Load` once or twice for an entire project, versus once per file in the legacy semantics.
2. Hierarchical configuration model that allows interface-specific config to be inherited from package-level config, which is inherited from defaults.
3. Single configuration file that describes the entirety of mockery's behavior, instead of spread out by `//go:generate` statements.
4. Extensive and flexible usage of a Golang string templating environment that allows users to dynamically specify parameter values.

Configuration Changes
----------------------

The existence of the `#!yaml packages:` map in your configuration acts as a feature flag that enables the feature.

The configuration parameters used in `packages` should be considered to have no relation to their meanings in the legacy scheme. It is recommended to wipe out all previous configuration and command-line parameters previously used.

The [configuration docs](/mockery/configuration/#packages-config) show the parameters that are available for use in the `packages` scheme. You should only use the parameters shown in this section. Mockery will not prevent you from using the legacy parameter set, but doing so will result in undefined behavior.

All of the parameters in the config section can be specified at the top level of the config file, which serves as the default values. The `packages` config section defines package-specific config. See some examples [here](/mockery/features/#examples).

`//go:generate` directives
----------------------------

Previously, the recommended way of generating mocks was to call `mockery` once per interface using `//go:generate`. Generating interface-specific mocks this way is no longer supported. You may still use `//go:generate` to call mockery, however it will generate all interfaces defined in your config file. There currently exists no semantics to specify the generation of specific interfaces from the command line (not because we reject the idea, but because it was not seen as a requirement for the initial iteration of `packages`).

Behavior Changes
-----------------

The legacy behavior iterated over every `.go` file in your project, called [`packages.Load`](https://pkg.go.dev/golang.org/x/tools/go/packages#Load) to parse the syntax tree, and generated mocks for every interface found in the file. The new behavior instead simply grabs the list of packages to load from the config file, or in the case of `#!yaml recursive: True`, walks the filesystem tree to discover the packages that exist (without actually parsing the files). Using this list, it calls `packages.Load` once with the list of packages that were discovered.

Filesystem Tree Layouts
------------------------

The legacy config provided the `keeptree` parameter which, if `#!yaml keeptree: True`, would place the mocks in the same package as your interfaces. Otherwise, it would place it in a separate directory.

These two layouts are supported in the `packages` scheme. See the relevant docs [here](/mockery/features/#layouts).
