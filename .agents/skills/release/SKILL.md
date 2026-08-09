---
name: release
description: >
  Cuts a versioned release of pulumi-garage (provider binary + SDKs). Use
  when asked to cut a release, tag a version, publish a new version, or
  explains the release process. Triggers the tag-push release flow and
  explains required registry secrets.
compatibility: Requires git and GitHub CLI (gh)
allowed-tools: Bash(git:*) Bash(gh:*)
---

# Release Skill

Releases are triggered by pushing a semver tag - there is no
`workflow_dispatch` release trigger in this repo. Never hand-build or
hand-publish an SDK for a release; always go through the tag-push workflow.
See [RELEASING.md](../../../RELEASING.md) for the full picture (this skill
is a condensed operational summary of it - if they ever disagree,
RELEASING.md is the source of truth).

## Pre-flight

1. Confirm `main` is green: `gh run list --branch main --limit 1`.
2. Confirm the secrets a release depends on are configured, or accept that
   the unconfigured ones will be skipped rather than fail (see RELEASING.md's
   Required secrets table: `PYPI_API_TOKEN` is the only stored secret; npm
   and NuGet use OIDC trusted publishing).

## Cutting a release

```sh
git tag v1.0.0          # or v1.0.0-alpha.1 for a prerelease
git push origin v1.0.0
```

Prerelease tags (any semver tag with a `-` suffix) are marked as a GitHub
prerelease automatically (`release.prerelease: auto` in `.goreleaser.yml`)
and published to npm under the `next` dist-tag instead of `latest`.

## What the tag push triggers

`.github/workflows/push.release.yaml`:

1. Builds the provider binary for darwin/linux/windows (amd64+arm64) via
   GoReleaser and publishes a GitHub Release with checksums - this alone
   is enough for `pulumi plugin install resource garage <version>` to work.
2. Pushes a second tag, `sdk/go/pulumi-garage/vX.Y.Z`, so the Go SDK
   resolves cleanly via `go get` despite living in a repo subdirectory.
3. Publishes the Node.js, Python, and .NET SDKs to their registries - each
   independently gated on its own secret/trusted-publisher setup; an
   unconfigured registry is skipped, not failed.

## Watching a release run

```sh
gh run list --repo axnic/pulumi-garage --workflow=push.release.yaml --limit 5
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
