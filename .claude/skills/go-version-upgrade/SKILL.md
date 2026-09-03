---
name: go-version-upgrade
description: Update mockery to support a new Go minor release, or backport that update to another branch. Covers which files get which version, why the root go directive lags the newest Go, upgrading golangci-lint alongside it, how to retidy the workspace without dependency churn, and how to verify against the new toolchain locally. Use when a new Go 1.N ships, when asked to "support/upgrade to Go 1.N", when a PR bumps go directives or the CI matrix, or when CI reports `internal error: package X without types was imported from Y` or mass `(typecheck)` failures in generated mocks.
---

# Upgrading mockery for a new Go release

## Support policy

mockery supports the **two most recent stable Go minor versions**. The authoritative
prose lives in `docs/dev-notes.md` (v3 branch) under "Go Version Support" and "Go
Syntax Updates" — read it first, and update it if the procedure changes.

The single most important rule: **the root `go.mod` `go` directive is NOT what adds
support for a new Go release.** It declares the minimum Go version required to build
mockery. Setting it to the newest Go locks out users still on the previous release.

Support for a new release comes from **upgrading the things that read Go type export
data** — `golang.org/x/tools` in the root module, and golangci-lint in `tools/`. Each
Go release can change the export data format (Go 1.27 moved to pkgbits V3/V4), and
anything carrying an older `x/tools` silently loses type information rather than
failing cleanly.

## What gets which version

For a new release Go `1.N` (so `1.(N-1)` is the older supported version):

| Location | Value | Why |
| --- | --- | --- |
| `golang.org/x/tools` in root `go.mod` | newest release supporting `1.N` | **The actual fix.** Loads/parses target packages via `go/packages`. |
| `golangci-lint` in `tools/go.mod` | version supporting `1.N` | **Also required.** Carries its own vendored `x/tools`; an old one fails typecheck on every generated mock. |
| root `go.mod` `go` directive | `1.(N-1).0` | Support floor. Never `1.N` — it would break the `1.(N-1)` CI job and every user on `1.(N-1)`. |
| fixture submodule `go.mod` | `1.(N-1).0` | Built by the same test matrix; keep it at the floor. |
| `tools/go.mod` | whatever `go mod tidy` demands | Dev-only, never published, so it may run ahead of the floor. golangci-lint tends to force it. |
| `go.work` `go` directive | `>=` every member module's directive | A Go requirement. Bump it when `tools/` rises. |
| CI matrix `go_vers` | `["1.(N-1)", "1.N"]` | Both supported versions. |
| `Dockerfile` | `golang:1.N-alpine` | The shipped image must parse `1.N` source, so it uses the *newest* Go — the opposite of the `go` directive. |

Example end state after the Go 1.27 upgrade: root `go.mod` `go 1.25.x` +
`x/tools v0.49.0`, `tools/go.mod` `go 1.26.0` + `golangci-lint/v2 v2.13.1`,
`go.work` `go 1.26.0`, matrix `["1.26", "1.27"]`, Dockerfile `golang:1.27-alpine`.

Do **not** touch `mockery-tools.env` / `VERSION` — that is the separate release flow.

## Branch layout

| | v3 (default) | v2 (maintenance) |
| --- | --- | --- |
| module | `github.com/vektra/mockery/v3` | `github.com/vektra/mockery/v2` |
| parser | `internal/pkg` | `pkg/parse.go` |
| fixture submodule | `internal/fixtures/example_project/pkg_with_submodules/go.mod` | `pkg/fixtures/example_project/pkg_with_submodules/go.mod` |
| CI matrix OSes | macos, ubuntu, windows | macos, ubuntu |
| dev notes | `docs/dev-notes.md` | not present — read v3's |

Do v3 first, then backport to v2. v2's `test.ci` still runs lint, so the golangci-lint
upgrade is not optional there even though v2 has no separate lint job.

## Procedure

1. **Bump `golang.org/x/tools`** in the root module:
   ```bash
   GOWORK=off GOFLAGS=-mod=mod go get golang.org/x/tools@vX.Y.Z
   ```
2. **Edit the go directives** — root `go.mod`, `go.work`, fixture submodule `go.mod` —
   per the table above. Drop any stale `toolchain` directive unless mockery actually
   needs a newer toolchain than its floor.
3. **Upgrade golangci-lint** in `tools/` (see below).
4. **Edit the CI matrix** in `.github/workflows/reusable-testing.yml` and the
   `FROM golang:` line in `Dockerfile`.
5. **Retidy each module separately, with the workspace off:**
   ```bash
   GOWORK=off GOFLAGS=-mod=mod go mod tidy          # repo root
   cd tools && GOWORK=off GOFLAGS=-mod=mod go mod tidy
   ```
   Then refresh `go.work.sum` by building in workspace mode:
   ```bash
   go build ./... && (cd tools && go build ./...)
   ```
   **Do not run `go work sync`.** It resolves the workspace-wide maximum for every
   shared dependency and writes those upgrades back into each module, producing a
   large unrelated diff in `tools/` (zerolog, x/crypto, x/net, …).
6. **Update `docs/dev-notes.md`** if the procedure itself changed.
7. **Verify** — see below.

## Upgrading golangci-lint

The version is pinned by a blank import in `tools/tools.go` plus the require in
`tools/go.mod`, and invoked via `go run` from the `lint` task in `Taskfile.yml`. All
three must agree; a major-version bump changes the module path.

```bash
# in tools/
sed -i '' 's|golangci-lint/cmd|golangci-lint/v2/cmd|' tools.go
GOWORK=off GOFLAGS=-mod=mod go get github.com/golangci/golangci-lint/v2@vX.Y.Z
GOWORK=off GOFLAGS=-mod=mod go mod tidy
# then update the lint cmd in Taskfile.yml to match the new path
```

For a v1 → v2 jump the config schema also changes. Convert it mechanically, then
review:

```bash
cd tools && GOWORK=off go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint \
  migrate -c ../.golangci.yml    # writes .golangci.yml, backs up .golangci.bck.yml
rm .golangci.bck.yml
```

Two things `migrate` gets wrong, both already solved in v3's `.golangci.yml` — copy
its handling:

- **staticcheck absorbed gosimple (`S*`), stylecheck (`ST*`) and quickfix (`QF*`).**
  `checks: [all, -SA1024]` therefore silently enables whole check families that were
  never on before. Disable `-ST*`, `-QF*` and the specific `S*` checks the old
  gosimple set omitted, or expect a wave of unrelated findings.
- **gofumpt moved from `linters` to `formatters`.**

Expect the newer gofumpt/testifylint to flag a few genuine pre-existing nits that the
old versions missed. Fix them rather than adding exclusions.

## Verifying

You do **not** need the new toolchain installed — `GOTOOLCHAIN` downloads it on
demand, and this is the only way to reproduce the failures these upgrades exist to
fix:

```bash
GOTOOLCHAIN=go1.N.P go version    # downloads if absent
GOTOOLCHAIN=go1.N.P go run github.com/go-task/task/v3/cmd/task test.ci
GOTOOLCHAIN=go1.(N-1).P go run github.com/go-task/task/v3/cmd/task test.ci
```

`test.ci` runs fmt and lint, then removes and regenerates every mock, then unit
tests, then `./e2e/run_all.sh`. Two traps when reading its output:

- **task fingerprints results.** `Task "test" is up to date` / `Task "lint" is up to
  date` means it was *skipped*, not that it passed. Check the exit code, and run
  `go test ./...` or `golangci-lint run` directly under each toolchain when you need
  a real signal.
- The `test_missing_interface` e2e case logs `Failed to run task "mocks.generate":
  exit status 1` **on purpose**. Judge the run by the overall exit code.

Then confirm regeneration is a no-op — `git status` should show no diff under
`mocks/` or in `*_mock.go` / `mock_*_test.go` fixtures. `go/types` and stdlib changes
between Go releases can legitimately alter generated identifiers and signatures even
with no template change, so investigate any diff rather than committing it blindly.

Finally, confirm the tree is otherwise clean. A stray `go.work.sum` hash is the usual
culprit, and on v3 the CI git-state check fails on it. That is why v3's workflow pins
`GOWORK: "off"` on its `go mod download` step: in workspace mode a bare `go mod
download` fetches the whole workspace build list, including graph-only deps of the
tools module, and records their hashes.

## Symptoms → causes

- **`internal error: package "fmt" without types was imported from "..."`** —
  root module's `golang.org/x/tools` is too old to read the new release's export data.
  Bump `x/tools` (issue #1171; v3 PR #1172 went v0.42.0 → v0.49.0).
- **Mass `_m.Called undefined (type *Foo has no field or method Called) (typecheck)`
  across `mocks/`** — same root cause, but inside golangci-lint: its vendored
  `x/tools` cannot read testify's export data, so the embedded `mock.Mock` resolves to
  nothing. Upgrade golangci-lint; do not touch the mock templates.
- **CI green on `1.N`, red on `1.(N-1)`** — the root `go` directive was set to `1.N`.
  Move it back to the floor.
- **`go.work` rejected / "requires go >= X"** — a member module's directive now
  exceeds `go.work`'s. Bump `go.work`.
- **Huge unrelated diff in `tools/go.mod` and `tools/go.sum`** — `go work sync` ran.
  `git checkout -- tools/go.mod tools/go.sum` and redo step 5.

### Beware the plausible-looking workaround

Adding `packages.NeedDeps` to the parser's `packages.Config` mode also makes the
"without types" error disappear, and a contributor PR may propose exactly that. It
works by type-checking the entire transitive dependency graph from source instead of
reading export data, roughly doubling package load time (measured 0.27s → 0.59s warm
on mockery's own config; worse on large trees). It treats the symptom. Bump `x/tools`
instead.

Prove which fix is load-bearing before committing either: revert the candidate
workaround, reproduce the original error under `GOTOOLCHAIN=go1.N.P`, then apply the
dependency bump alone and confirm it clears.

## Conventions

- Branch: `go1NN_v2` / `go1NN_v3` (e.g. `go127_v2`).
- Commit subjects: `v2: Update for Go 1.N` for the version/matrix move,
  `fix(ci): ...` when the change is confined to CI and the tools module.
- Reference the issue: `Fixes #NNNN` on the commit that closes it, `Refs #NNNN` on
  follow-ups.
- When implementing or porting someone else's PR rather than merging it, credit them:
  `Co-Authored-By: Name <email>`, taken from `gh pr view <n> --json commits`.
