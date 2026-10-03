# Tech Documentation Analysis

## Project
- Last Commit: 9962f71f75e50e229f84563268570ea3c0497019
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
- CI is central: hand-written workflows were removed; the callers in
  `.github/workflows/` (Terraform-generated, thin) call reusable workflows from
  `axnic/.github`: `qa` (Quality Assurance: lint via `rtunk`, commit messages),
  `test` (build, tests, coverage), `scan`, `deps`, `audit`, `codegen`, `release`,
  `e2e-sync`. Triggers on merge_group, pull_request, AND push to main (commits land
  on main directly, without a PR, so the push trigger is what actually validates
  them). Don't re-add logic to this repo's workflows: CI calls `mise run <task>`
  (`lint`, `ci:commitlint`, `ci:build`, `ci:test`, `ci:coverage`, `security:audit`,
  `ci:e2e`, `ci:e2e:versions`; see `.config/mise.toml`), and `mise run ci` is the
  local aggregate. `mise run lint` = `rtunk check .` (config: `.rtunk/rtunk.yaml`),
  not `make lint`.
- Coverage gate (`mise run ci:coverage`): reads `provider/coverage.txt` (written by
  `make test`), runs `go tool cover -func` for total statement coverage, and fails
  below 60%. This is a floor/ratchet, not a target - raise it as coverage improves,
  never lower it just to unblock a failing PR.
- Branch protection on `main` (a live GitHub repo setting, not a file in this
  repo) requires 4 status checks — Lint, Commit Messages, Build, Tests —
  before a PR/merge-queue merge. It does NOT gate direct `git push` to main
  (classic branch-protection required-status-checks only covers PR/merge-queue
  merges, not raw pushes), so the direct-push-to-main workflow described above
  is unaffected.
- Dependabot (`.github/dependabot.yml`): routine weekly version-update PRs for
  github-actions, gomod and npm (/examples); what happens to them afterwards
  (approval, auto-merge) is handled by the central Dependency Updates workflow in
  `axnic/.github`, not by a file in this repo. Dependabot's own PR titles are never
  sentence-case and use scope `deps-dev` for dev dependencies (not in
  `.commitlintrc.js`'s scope-enum, only `deps` is allowed), so no `commit-message:`
  config is set here and merge automation must build a compliant subject. "Allow
  auto-merge" and Discussions are live repo settings (Discussions: linked from
  `.github/ISSUE_TEMPLATE/config.yml`).
- `make test_all` was removed from the Makefile — it referenced
  `provider/pkg` and `tests/sdk/{nodejs,python,dotnet,go}`, none of which
  exist in this repo (leftover from the `pulumi-resource-provider-boilerplate`
  template this repo was generated from). Never called by CI or documented in
  CONTRIBUTING.md.
- `.agents/skills/`: 3 of 6 skill files were found to be stale copies from an
  unrelated sibling project (`pi-extension-settings`, a Node.js/npm tool with
  a different scope-enum, different repo owner
  `xunleii/pi-extension-settings`, different release mechanism) and have been
  corrected — don't rediscover this: `open-pr/SKILL.md` (wrong
  scopes/check-commands/doc filenames/branch-naming table — now correct,
  defers to `commit/SKILL.md` for commit-message rules instead of duplicating
  them, and documents the 4 required branch-protection checks above);
  `release/SKILL.md` (releases are a `workflow_dispatch` of the central Release workflow with
  `bump` XOR `version`, not a tag push — points to RELEASING.md as the source of truth); and
  `create-issue/SKILL.md` + `references/templates.md` +
  `references/examples.md` (two hardcoded URLs pointed at
  `xunleii/pi-extension-settings`; bug-report required-fields table and
  templates/examples rewritten to match this repo's actual
  `.github/ISSUE_TEMPLATE/bug.yaml`/`feature.yml`, which were already
  correct — only the skill's docs about them were wrong). `release-notes/`
  (`SKILL.md` + `references/examples.md`) was deleted entirely — it described
  a "deterministic draft + Copilot polish" release-notes pipeline that
  doesn't exist anywhere in this repo (`.goreleaser.yml` uses `changelog:
  use: github`, GitHub's native auto-generated notes), so it was pure
  fiction for this project. `commit/SKILL.md` and
  `pulumi-provider-review/SKILL.md` were checked and found already accurate
  — no changes made there.
- Compatibility matrix: one thin caller per Garage version,
  `.github/workflows/merge_group,pull_request,push.e2e-v2.{0,1,2,3}.0.yaml`, each
  calling the central E2E workflow in `axnic/.github` with a pinned version. One file
  per version rather than a `strategy.matrix` job because GitHub Actions status badges
  are per-workflow-file, not per-matrix-leg - the README's Compatibility table embeds
  each file's own badge. New versions are added by the E2E Sync workflow of
  `axnic/.github`, which opens a PR with the new caller and the README badge row
  (keep the row format `| vX.Y.0 | [![E2E (Garage vX.Y.0)](...)](...) |`). Confirm
  `docker-compose.yml` / `scripts/bootstrap-garage.sh` still work against a new version
  (`GARAGE_VERSION=vX.Y.Z make test_e2e`) - `--single-node` only exists from Garage
  v2.3.0 onward, so the bootstrap script always does the manual layout bootstrap,
  which is what makes this version-agnostic.
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
  - Gotcha: a fresh build's `postCreateCommand` failed outright with `mise
    ERROR Failed to install vfox:version-fox/vfox-dotnet@8.0.20: run with
    --yes to install plugin automatically` - mise refuses to auto-install a
    *new* asdf/vfox plugin without explicit consent. GitHub Actions CI gets
    this for free (it sets `CI=true`, which mise treats as consent) but
    nothing set that inside the devcontainer. Fixed by adding `MISE_YES:
    "1"` to the `dev` service's `environment:` block in
    `.devcontainer/docker-compose.yml`.
  - Gotcha: even after `postCreateCommand` succeeded, a plain terminal/exec
    session inside the container had NO mise-installed tools on PATH at all
    (`go`, `pulumi`, `golangci-lint`, etc. all "command not found") -
    `make lint` and `make test` did NOT "work immediately" as
    CONTRIBUTING.md claims. The `ghcr.io/devcontainers-extra/features/mise:1`
    feature only installs the `mise` binary itself (confirmed by reading its
    install.sh) - it configures no shell activation or PATH wiring, and this
    repo's config never added any either; it only ever worked, if it did, via
    the `hverlin.mise-vscode` VS Code extension patching PATH for its own
    integrated terminal specifically - not for a plain `docker exec`, a VS
    Code task, another editor, or Codespaces. Fixed with a `remoteEnv` block
    in `.devcontainer/devcontainer.json` appending
    `/home/vscode/.local/share/mise/shims` to PATH. `containerEnv` (not
    `remoteEnv`) was tried first and crash-looped the container (`sleep: not
    found`) - for a compose-based devcontainer, `containerEnv`'s
    `${containerEnv:PATH}` substitution resolves empty at container-creation
    time and clobbers the base PATH entirely; `remoteEnv` applies only to the
    CLI's managed remote processes (terminals/exec/postCreate) and doesn't
    touch container boot, which is why it's the correct one to use here.
  - Verified end-to-end with the real `@devcontainers/cli` (`npx
    @devcontainers/cli up --workspace-folder .`), not just read for
    plausibility - caught all four gotchas above by actually building and
    running it. A full fresh build followed by `make lint` (0 issues) and
    `make test` (including the example-lifecycle E2E tests against the real
    always-on Garage instance, per CONTRIBUTING.md's documented devcontainer
    behavior) were verified passing end-to-end after both new fixes.
