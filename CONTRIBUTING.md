# Contributing to pulumi-garage

Thanks for your interest in contributing. This document covers setting up a dev
environment, the test-driven workflow this project follows, generating the SDKs, commit
conventions, and how a release gets cut.

## Code of Conduct

Please read our [Code of Conduct](CODE-OF-CONDUCT.md) before participating.

## Setting up your development environment

### Devcontainer (recommended)

Open the repo in VS Code with the
[Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)
(or a GitHub Codespace) and "Reopen in Container" - everything is ready the moment it
finishes building, no manual setup step:

- Go, the Pulumi CLI, and everything else pinned in `.config/mise.toml` are installed
  via the [mise devcontainer feature](https://github.com/devcontainers-extra/features/tree/main/src/mise)
  (`ghcr.io/devcontainers-extra/features/mise:1`) and `postCreateCommand`.
- A single-node Garage instance (`.devcontainer/docker-compose.yml`) is already
  running, sharing this container's network namespace (`network_mode: service:garage`)
  and already bootstrapped, so `http://localhost:3903` / `http://localhost:3900` just
  work.
- `GARAGE_ADMIN_ENDPOINT` / `GARAGE_ADMIN_TOKEN` are already exported (see the `dev`
  service's `environment:` block), and a starter S3 access key named `dev` is created
  on first boot - its credentials are printed once in the "Reopen in Container" log
  (`GARAGE_DEV_ACCESS_KEY_ID` / `GARAGE_DEV_SECRET_ACCESS_KEY`). It isn't granted any
  bucket permissions yet; wire it into a `BucketKeyPermission` as part of whatever
  you're testing.

`make lint` and `make test` work immediately - and because `GARAGE_ADMIN_ENDPOINT` is
already set, `make test` runs the example lifecycle tests too (normally skipped
without a live cluster), so it's a fuller check inside the devcontainer than outside
it. `make dev-up` / `make test_e2e` also work, via a *separate*, disposable Garage
stack (see below) reached through the `docker-outside-of-docker` feature - no port
conflict with the always-on instance, since that one publishes no host ports of its
own.

### Without a devcontainer

Any environment with Go, the Pulumi CLI, and Docker works (see `.config/mise.toml` for
exact pinned versions - `mise install` picks all of them up automatically):

```sh
make dev-up    # starts Garage and bootstraps its single-node layout, prints
               # GARAGE_ADMIN_ENDPOINT / GARAGE_ADMIN_TOKEN to export
pulumi up      # or: cd examples/yaml && pulumi up
make dev-down  # tear it down when you're done
```

`make dev-up` is `make test_e2e`'s setup half, minus the automated test run and
teardown - the cluster stays up until you `make dev-down` it. Pin a version the same
way: `GARAGE_VERSION=v2.0.0 make dev-up`.

## Test-driven development

This project is developed test-first: a failing test before the implementation that
makes it pass, for every resource and every client method. When adding or changing
behavior, follow the same pattern - it's not a style preference here, it's how the
existing code was built and how regressions get caught.

Tests are layered, from fastest/most-isolated to slowest/most-realistic:

1. **Admin API client** (`provider/internal/garageclient/*_test.go`) - against a real
   `httptest.Server` replaying exact Admin API v2 JSON responses.
2. **Resource logic** (`provider/*_resource_test.go`) - against a minimal in-memory
   `garageAPI` stub, no HTTP involved, to test Create/Read/Update/Delete/Diff quickly
   and deterministically.
3. **Provider wiring** (`provider/*_test.go`) - integration tests via
   `pulumi-go-provider`'s test harness, to verify schema/config/lifecycle wiring
   end-to-end without an external dependency.
4. **Example-program lifecycle** (`examples/*_test.go`) - via `providertest`/
   `pulumitest`, exercising the real provider against a real Garage cluster. Skipped
   (`t.Skip`) when `GARAGE_ADMIN_ENDPOINT` isn't set, so they run automatically inside
   the devcontainer but don't slow down a plain `make test` elsewhere.

Commands:

- `make test` - fast, hermetic unit tests (layers 1-3, plus layer 4 wherever a live
  cluster happens to be reachable).
- `make test_e2e` - spins up a real, disposable, single-node Garage cluster via Docker
  Compose (`docker-compose.yml`), runs the example programs' full create/update/delete
  lifecycle against it plus a real S3 `PutObject`/`GetObject` round trip, then tears it
  down. Requires Docker. Defaults to `dxflrs/garage:v2.3.0`; pin another version with
  `GARAGE_VERSION`, e.g. `GARAGE_VERSION=v2.0.0 make test_e2e` - see the
  [Compatibility matrix](README.md#compatibility) for which versions CI verifies.
- `make lint` - runs `golangci-lint`.

CI (`.github/workflows/merge_group,pull_request,push.ci.yaml`) runs lint, commitlint,
build, and unit tests on every pull request, merge-queue entry, and push to `main` -
this repo takes commits directly on `main` without a PR for most changes, so the push
trigger is what actually validates them. The E2E suite runs separately, once per
supported Garage version - see the [Compatibility matrix](README.md#compatibility).

## Building and generating the SDKs

Build and install the provider binary:

```sh
make build install
```

After changing resource code (fields, descriptions, new resources), regenerate the
schema (`provider/cmd/pulumi-resource-garage/schema.json`) and all 5 SDKs:

```sh
make codegen
```

This regenerates `sdk/{go,nodejs,python,dotnet,java}/` from the schema. Only the
generated *source* under `sdk/*/` is committed; per-language build/publish output
(`sdk/python/venv/`, `sdk/nodejs/bin/`, `sdk/java/.gradle/`, ...) is gitignored (see the
root `.gitignore` - the per-language `.gitignore`/`.gitattributes` files that codegen
itself generates get wiped on every run, so they can't be hand-edited to cover this).

To build a single language's SDK locally (e.g. to smoke-test a packaging change),
`make {nodejs,python,dotnet,java,go}_sdk` runs codegen plus that language's own build
step (`yarn install && tsc`, a Python venv + `build`, `dotnet build`, `gradle build`).

Field-level documentation comes from `infer.Annotate` calls in each resource's
`*_resource.go` (see `provider/bucket_resource.go` for the pattern) - these flow
straight into `schema.json` and from there into every generated SDK's doc comments, so
document there, not by hand-editing generated SDK output.

## Commit conventions

Commits follow [Conventional Commits](https://www.conventionalcommits.org/), enforced
by commitlint (`.commitlintrc.js`) in CI:

```
type(scope): Subject starting with uppercase

Body — one or more paragraphs explaining WHY, wrapped at 80 chars/line.

Signed-off-by: Name <email>
```

- **Type** - one of `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`,
  `build`, `ci`, `chore`, `revert`.
- **Scope** - exactly one of `provider`, `sdk`, `examples`, `tests`, `docs`, `ci`,
  `deps`, `tooling` (see `.commitlintrc.js` for the full list with descriptions).
  Required, and only one per commit - a commit touching two unrelated areas for two
  unrelated reasons should be two commits.
- **Subject** - sentence case, no trailing period, ≤100 chars.
- **Body** - required, explains *why*, not just what (the diff already shows what).

Run `npx commitlint --edit` (or let the `commit-msg` hook do it) to validate a message
before committing. See `.agents/skills/commit/SKILL.md` for the full convention this
project's AI assistant follows, including trailers.

## Repo layout

- `provider/` - the provider implementation (`provider.go`, `config.go`,
  `*_resource.go`), plus the internal Garage Admin API client
  (`provider/internal/garageclient/`) and the codegen entrypoint
  (`provider/cmd/pulumi-resource-garage/`).
- `sdk/` - the generated SDKs for all 5 languages (`make codegen`).
- `examples/` - the YAML and Go example programs, and their lifecycle tests.
- `docker-compose.yml`, `test/e2e/garage.toml`, `scripts/bootstrap-garage.sh` - the
  local dev / CI E2E fixture: a disposable single-node Garage cluster with a fixed,
  test-only `rpc_secret`/`admin_token` (not sensitive - the cluster is ephemeral and
  local-only), and the script that bootstraps its layout.
- `.devcontainer/` - VS Code / Codespaces dev environment (its own
  `docker-compose.yml`, an always-on Garage instance), see
  [Setting up your development environment](#setting-up-your-development-environment).
- `Makefile` - build, codegen, lint, test, and dev-cluster targets.

## Releasing

Publishing the provider binary and the 5 SDKs to their registries is a maintainer
task - see [RELEASING.md](RELEASING.md) for the required secrets and how to cut a
release. You don't need any of this to build, test, or contribute.
