# `#!yaml resolve-internal-aliases:`

Some libraries declare their API in an `internal` package and re-export it with type aliases:

```go title="client/client.go"
package client

import "example.com/lib/internal"

type StartWorkflowOptions = internal.StartWorkflowOptions
```

Method signatures are written against the original type, so a mock of such an interface would import `example.com/lib/internal`, which Go forbids outside that module. When `#!yaml resolve-internal-aliases: true`, mockery instead uses the public alias (`client.StartWorkflowOptions`).

=== "before"

    ```go
    func (_mock *Client) ExecuteWorkflow(ctx context.Context, options internal.StartWorkflowOptions) error
    ```

=== "after"

    ```go
    func (_mock *Client) ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions) error
    ```

## Configuration

This is a root-level setting. Either every package is searched for aliases or none are.

```yaml title=".mockery.yml"
resolve-internal-aliases: true
```

## How it works

While loading the packages in your config, mockery records every exported alias whose target is a named type in an `internal` package. Method parameters and results are then rewritten using that table, but only for types the output package is not allowed to import (so mocks generated inside the library's own tree are left alone), including types nested in pointers, slices, arrays, maps, channels, generic arguments, func signatures, struct and interface literals and unions. If several packages alias the same type, the alias with the smallest name (e.g. `example.com/lib/client.Foo`) is used.

This is the automatic counterpart to [`replace-type`](replace-type.md), which maps each type by hand.

## Limitations

- Aliases are only found in packages that mockery loads, so list the packages that do the re-exporting.
- Type-parameter constraints of generic interfaces are not rewritten.
- Types with no public alias cause mockery to fail with an error naming the interface, method and type, because the generated code could not compile.
- Generic types that are themselves declared in an internal package (as opposed to generic types with internal arguments, which are handled) are not supported, as they would need generic aliases.
