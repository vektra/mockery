---
title: Third-Party
---

Mockery's template system lets anyone write and distribute their own templates. The templates below are maintained outside of the mockery project. They can be used through the [`#!yaml template: "file://"`](index.md#template-options) or [`#!yaml template: "https://"`](index.md#template-options) options.

!!! warning

    Third-party templates are not maintained or reviewed by the mockery project. Check each project's documentation, license, and release history before using it, and pin a specific version.

Want your template listed here? Open a pull request that adds it to this page.

## [`powerman/mockery-templates`](https://github.com/powerman/mockery-templates)

[`powerman/mockery-templates`](https://github.com/powerman/mockery-templates) is a collection of templates that generate **stateful fakes**. Each method of a fake has a default hook, a FIFO queue of one-shot hooks, and a call history. Unexpected calls panic, so a test can't quietly pass on zero values. The generated code only depends on the standard library.

| Template | Description |
|----------|-------------|
| [`fake`](https://pkg.go.dev/github.com/powerman/mockery-templates/fake) | Stateful fakes with their own API, which can change between releases. Derived from go-mockgen. Recommended by the author. |
| [`go-mockgen`](https://pkg.go.dev/github.com/powerman/mockery-templates/go-mockgen) | Stateful fakes with the same API as [go-mockgen](https://github.com/derision-test/go-mockgen) |

The generated API, naming rules, and `.mockery.yml` examples for each template are on its pkg.go.dev page. The project's [RATIONALE.md](https://github.com/powerman/mockery-templates/blob/main/RATIONALE.md) explains why it builds on mockery instead of a dedicated generator.

### `fake` example

`fake` panics on any call that has no hook configured. Call records have fields named after the method's parameters and results, so `Get(ctx, key)` is recorded as `call.Key`.

=== "Interface"

    ```go
    package port

    type Store interface {
        Get(ctx context.Context, key string) (value string, err error)
        Set(ctx context.Context, key, value string) error
    }
    ```

=== "Example Usage"

    ```go
    func TestStoreFake(t *testing.T) {
        ctx := context.Background()

        // Default hooks call a real (e.g. in-memory) implementation.
        store := porttest.NewFakeStoreFrom(newMemStore())
        // Only the next call to Set fails; later calls go back to the default hook.
        store.SetFunc.PushReturn(errDiskFull)

        assert.ErrorIs(t, store.Set(ctx, "key", "one"), errDiskFull)
        assert.NoError(t, store.Set(ctx, "key", "two"))

        value, err := store.Get(ctx, "key")
        assert.NoError(t, err)
        assert.Equal(t, "two", value)

        // Call records have fields named after parameters and results.
        calls := store.GetFunc.History()
        assert.Len(t, calls, 1)
        assert.Equal(t, "key", calls[0].Key)
        assert.Equal(t, "two", calls[0].Value)

        // NewFakeStore has no default hooks: calls panic unless configured.
        strict := porttest.NewFakeStore()
        strict.GetFunc.PushReturn("once", nil)
        _, _ = strict.Get(ctx, "key")
        assert.PanicsWithValue(t, "unexpected invocation of FakeStore.Get", func() {
            _, _ = strict.Get(ctx, "key")
        })
    }
    ```

=== "`.mockery.yml`"

    ```yaml
    _anchors:
      fake: &fake
        template: file://.mockery/fake.templ
        dir: '{{.InterfaceDirRelative}}/{{.SrcPackageName}}test'
        pkgname: '{{.SrcPackageName}}test'
        filename: 'fake.{{.InterfaceFile | base}}'
        structname: 'Fake{{.InterfaceName | firstUpper}}'
    packages:
      example.com/app/port:
        interfaces:
          Store:
            config: *fake
    ```

### `go-mockgen` example

`go-mockgen` generates the same API as go-mockgen, so existing tests keep working. `NewMockStore` returns zero values for calls with no hook configured, and `NewStrictMockStore` panics instead. Call records use positional `ArgN` and `ResultN` fields. The `structname` below matches the names that `go-mockgen --prefix Stub -i Store` would produce.

=== "Interface"

    ```go
    package port

    type Store interface {
        Get(ctx context.Context, key string) (string, error)
        Set(ctx context.Context, key, value string) error
    }
    ```

=== "Example Usage"

    ```go
    func TestStoreMock(t *testing.T) {
        ctx := context.Background()

        // NewMockStubStore returns zero values for calls with no hook configured.
        store := porttest.NewMockStubStore()
        store.GetFunc.SetDefaultReturn("", errNotFound)
        // One-shot hooks run first, in order, then the default hook takes over.
        store.GetFunc.PushReturn("first", nil)
        store.GetFunc.PushHook(func(_ context.Context, key string) (string, error) {
            return "second for " + key, nil
        })

        value, _ := store.Get(ctx, "key")
        assert.Equal(t, "first", value)
        value, _ = store.Get(ctx, "key")
        assert.Equal(t, "second for key", value)
        _, err := store.Get(ctx, "key")
        assert.ErrorIs(t, err, errNotFound)

        assert.NoError(t, store.Set(ctx, "key", "value")) // Not configured: zero values.

        // Call records use positional fields: Arg0 is ctx, Arg1 is key.
        calls := store.GetFunc.History()
        assert.Len(t, calls, 3)
        assert.Equal(t, "key", calls[0].Arg1)
        assert.Equal(t, "first", calls[0].Result0)

        // NewStrictMockStubStore panics on calls with no hook configured.
        strict := porttest.NewStrictMockStubStore()
        assert.Panics(t, func() { _ = strict.Set(ctx, "key", "value") })
    }
    ```

=== "`.mockery.yml`"

    ```yaml
    _anchors:
      stub: &stub
        template: file://.mockery/go-mockgen.templ
        dir: '{{.InterfaceDirRelative}}/{{.SrcPackageName}}test'
        pkgname: '{{.SrcPackageName}}test'
        filename: 'stub.{{.InterfaceFile | base}}'
        structname: 'MockStub{{.InterfaceName | firstUpper}}'
    packages:
      example.com/app/port:
        interfaces:
          Store:
            config: *stub
    ```

### Installation

The author recommends copying the template and its JSON schema into your project, for example into `.mockery/`:

```sh
t=fake # or go-mockgen
mkdir -p .mockery
for f in "$t.templ" "$t.templ.schema.json"; do
    curl -fsSLo ".mockery/$f" \
        "https://raw.githubusercontent.com/powerman/mockery-templates/refs/tags/v0.2.0/$t/$f"
done
```

Then point to the copy with `#!yaml template: file://.mockery/fake.templ`. You can also point to the template by URL. Mockery downloads the template, and the schema next to it, on every run:

```yaml
template: https://raw.githubusercontent.com/powerman/mockery-templates/refs/tags/v0.2.0/fake/fake.templ
```

Check the [releases page](https://github.com/powerman/mockery-templates/releases) for the latest version.
