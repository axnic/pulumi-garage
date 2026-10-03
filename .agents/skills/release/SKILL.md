---
name: release
description: >
  Cuts a versioned release of pulumi-garage (provider binary + SDKs). Use
  when asked to cut a release, tag a version, publish a new version, or
  explains the release process. Triggers the dispatch-based release flow
  and explains registry secrets.
compatibility: Requires git and GitHub CLI (gh)
allowed-tools: Bash(git:*) Bash(gh:*)
---

# Release Skill

Releases are cut by dispatching the `Release` workflow - pushing a tag does
not publish anything. Never hand-build or hand-publish an SDK for a release.
See [RELEASING.md](../../../RELEASING.md) for the full picture (this skill
is a condensed operational summary of it - if they ever disagree,
RELEASING.md is the source of truth).

## Pre-flight

1. Confirm `main` is green: `gh run list --branch main --limit 1`.
2. Confirm the registry secrets you need are configured (`NPM_TOKEN`,
   `PYPI_API_TOKEN`, `NUGET_API_KEY`; each optional per registry - see
   RELEASING.md). Setup details live in the `axnic/.github` wiki.

## Cutting a release

Provide `bump` (`auto|patch|minor|major`) **or** `version`, never both:

```sh
gh workflow run workflow_dispatch.release.yaml -f bump=auto
gh workflow run workflow_dispatch.release.yaml -f version=1.2.0
gh workflow run workflow_dispatch.release.yaml -f version=1.2.0-rc.1   # release candidate
```

Prerelease versions (any version with a `-` suffix) are marked as a GitHub
prerelease automatically (`release.prerelease: auto` in `.goreleaser.yml`).

## What the dispatch triggers

The central Release workflow (`axnic/.github`), called from
`.github/workflows/workflow_dispatch.release.yaml`:

1. Builds the provider binary for darwin/linux/windows (amd64+arm64) via
   GoReleaser and publishes a GitHub Release with checksums - this alone
   is enough for `pulumi plugin install resource garage <version>` to work.
2. Pushes a second tag, `sdk/go/pulumi-garage/vX.Y.Z`, so the Go SDK
   resolves cleanly via `go get` despite living in a repo subdirectory.
3. Publishes the Node.js, Python, and .NET SDKs to their registries,
   depending on which registry secrets are configured.

## Watching a release run

```sh
gh run list --repo axnic/pulumi-garage --workflow=workflow_dispatch.release.yaml --limit 5
gh run view <run-id> --repo axnic/pulumi-garage
```

## Verifying afterward

```sh
pulumi plugin install resource garage <version>
```

Then check whichever registries (npmjs.com, pypi.org, nuget.org) you expect
this release to have reached.

## Retrying a failed publish

Not all registries tolerate a blind retry - see RELEASING.md's "Retrying a
failed publish" section. In short: the GitHub Release and Go SDK tag steps
are safe to re-run; check npm/PyPI's own state before retrying those (NuGet's
`--skip-duplicate` is more forgiving).
