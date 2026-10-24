## resolve-internal-aliases

Mimics a library that declares its API in an `internal` package and re-exports it from a public one,
like `go.temporal.io/sdk/client` does:

- `lib/internal` declares `Service`, `Options` and `Result`.
- `lib` re-exports them as type aliases.
- `mocks` is generated from `lib.Service`. It sits outside the `lib` tree, so Go does not let it import `lib/internal`.

The mock only compiles because `resolve-internal-aliases: true` makes mockery use `lib.Options` and `lib.Result`
instead of the internal types. See `mocks/service_test.go`.
