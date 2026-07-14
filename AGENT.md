# Tech Documentation Analysis

## Project
- Last Commit: 33f1e4d69fd6ea9139c56f28b3a20911abb93910
- Stack: Go 1.25 (pulumi-go-provider v1.1.2), Pulumi CLI 3.206.0, Docker (E2E only)
- Doc Language: en
- Module: github.com/axnic/pulumi-garage
- Provider namespace: axnic; resource tokens: garage:index:Bucket, garage:index:Key, garage:index:BucketKeyPermission

## Doc Locations
- Root entry point: README.md (single comprehensive doc, no sub-files by design for v1)
- Example-local doc: examples/yaml/README.md (smaller, mirrors quickstart config steps)
- No docs/adr, docs/api, or docs/guides directories exist in this repo

## Style by Type
| Type | Format | Tone | Example File |
|------|--------|------|--------------|
| README | H2 sections: Prerequisites/Quickstart/Resource reference/Provider config/Dev & testing/Compatibility/Local development/Known limitations/Additional details | Terse, no marketing fluff, no emojis | README.md |
| Go doc comments (Annotate) | Full sentences, authoritative field descriptions, quoted verbatim in README | Precise/technical | provider/*.go Annotate() methods |
| Makefile/CI comments | Short imperative comments above targets | Terse | Makefile, .github/workflows/*.yaml |

## Notes for future updates
- Ground truth for resource field descriptions lives in `Annotate()` methods in
  provider/{bucket,key,bucket_key_permission}_resource.go and provider/config.go —
  always re-read these directly (native Read, not a lossy/compressed reader) before
  updating the Resource reference or Provider configuration sections, since the
  exact wording is quoted in the README.
- Known v1 scope limits (cluster layout mgmt out of scope, single global alias only,
  Key.createBucket not modelled, non-empty Bucket delete fails, only yaml/go SDKs
  generated) are deliberate — keep the "Known limitations" section explicit, don't
  bury or soften it.
- nodejs/python/dotnet SDK dirs were deleted from sdk/ and examples/ as part of
  the boilerplate->real-provider conversion, then regenerated later for registry
  publishing (see git history) - don't reintroduce references to them without
  checking they've actually been regenerated. Java/Maven was regenerated once
  too, then dropped for good (low pulumi+java adoption vs. setup cost) - don't
  reintroduce it without an explicit ask.
- CI job list (as of last check): lint, commitlint, build, test — defined in
  .github/workflows/merge_group,pull_request,push.ci.yaml. Triggers on
  merge_group, pull_request, AND push to main (commits land on main directly,
  without a PR, so the push trigger is what actually validates them). E2E moved
  out of this file into a per-Garage-version matrix — see next point.
- Compatibility matrix (added for the version-matrix/devcontainer PR): four
  thin workflow files, `.github/workflows/merge_group,pull_request,push.e2e-garage-2.{0,1,2,3}.yaml`,
  each calling `.github/workflows/_reusable-e2e.yaml` (a `workflow_call`
  workflow) with a pinned `garage-version`. One file per version rather than a
  `strategy.matrix` job because GitHub Actions status badges are per-workflow-file,
  not per-matrix-leg — the README's Compatibility table embeds each file's own
  badge. If a version is added/removed, update: the matrix table in README.md,
  the four (or N) thin workflow files, and confirm `docker-compose.yml` /
  `scripts/bootstrap-garage.sh` still work against it (`GARAGE_VERSION=vX.Y.Z
  make test_e2e`) — `--single-node` only exists from Garage v2.3.0 onward, so
  the bootstrap script always does the manual layout bootstrap, which is what
  actually makes this version-agnostic. Gotcha hit and fixed: a job whose only
  content is `uses: ./.github/workflows/_reusable-e2e.yaml` needs its own
  explicit `permissions:` block (even matching the top-level `permissions: {}`
  in intent) - without one, all four workflows failed with `startup_failure`
  and zero jobs created, no useful error message anywhere in the GitHub API.
- `scripts/bootstrap-garage.sh` bootstraps a single-node layout purely over
  the Admin API (`GetClusterStatus` → `UpdateClusterLayout` →
  `ApplyClusterLayout`, all HTTP, driven by `GARAGE_ADMIN_ENDPOINT`/
  `GARAGE_ADMIN_TOKEN`) - NOT `docker exec`/`docker compose exec` (an earlier
  version did that, but it breaks as soon as the script runs somewhere that
  isn't the same `docker compose` invocation/project, e.g. from inside the
  devcontainer). Gotcha: Garage's Admin API returns pretty-printed JSON
  (`"key": "value"`, space after the colon) - naive `grep -o '"key":"[^"]*"'`
  silently extracts nothing; the `json_string_field`/`json_number_field`
  helpers in this script (and the matching one in
  `scripts/ensure-dev-key.sh`) tolerate the space. If you add a new script
  parsing Admin API JSON, reuse that pattern, don't re-derive it wrong.
- Local dev: `docker-compose.yml` (renamed from docker-compose.e2e.yml — it now
  serves both `make test_e2e` and interactive dev via `make dev-up`/`make
  dev-down`) plus `.devcontainer/` for a zero-setup devcontainer/Codespace:
  - Tools come from the official `ghcr.io/devcontainers-extra/features/mise:1`
    feature (not a hand-rolled Dockerfile install) - confirmed via
    `mise generate devcontainer`'s own reference output. `postCreateCommand`
    still runs `mise trust && mise install` against this repo's own mise
    config.
  - `.devcontainer/docker-compose.yml` defines its own always-on `garage`
    service plus a `dev` service on `network_mode: service:garage` (shares
    its network namespace, so `localhost:3903`/`:3900` just work inside the
    container - no `garage:3903` hostname to teach). `GARAGE_ADMIN_ENDPOINT`/
    `GARAGE_ADMIN_TOKEN` are set via that service's `environment:` block, so
    they're already exported in every shell.
  - `postCreateCommand` (`.devcontainer/postCreate.sh`) bootstraps the
    layout and runs `scripts/ensure-dev-key.sh` to mint a starter `dev` S3
    key, idempotently - see README's Local development section for the
    user-facing description.
  - Gotcha: `.config/mise.toml` pulls in the vfox-pulumi community plugin,
    which mise re-resolves on every invocation once the workspace is
    mounted — including a first `mise install` itself — so it has to be
    installed once, from a directory with no mise config in scope
    (`cd /tmp && mise plugins install vfox-pulumi ...`, in postCreate.sh),
    or every `mise` command in the workspace fails with "plugin not
    installed" no matter what you run.
  - Gotcha: `provider/config_test.go`'s `TestConfigureFailsWithout*` tests
    must `t.Setenv(envAdminEndpoint, "")` / `t.Setenv(envAdminToken, "")`
    explicitly - the devcontainer sets both ambiently for its always-on
    Garage, which silently turns the "neither set" negative test into a
    false pass/fail depending on where it runs, if you don't clear them.
  - Side effect worth knowing, not a bug: inside the devcontainer, `make
    test` also runs the example-lifecycle E2E tests (normally skipped via
    `requireGarage(t)`'s `t.Skip`), because `GARAGE_ADMIN_ENDPOINT` happens
    to already be set there. Documented in README as a feature, not
    "flakiness".
  - Verified end-to-end with the real `@devcontainers/cli` (`npx
    @devcontainers/cli up --workspace-folder .`), not just read for
    plausibility - caught both gotchas above by actually building and
    running it, twice.
