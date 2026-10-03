# Releasing

This document is for maintainers cutting a release of the provider and its SDKs.

## How a release works

A release is a manual dispatch of the `Release` workflow (caller
`.github/workflows/workflow_dispatch.release.yaml`, a thin caller of the central
Release workflow in `axnic/.github`). **Pushing a tag no longer publishes
anything.** The dispatch takes one of two inputs, mutually exclusive:

- `bump`: `auto` | `patch` | `minor` | `major` - the next version is computed
  from the latest release.
- `version`: an explicit version such as `1.2.0`, or a release candidate such
  as `1.2.0-rc.1`.

Setting both, or neither, is an error. The release:

1. Builds the provider binary for darwin/linux/windows (amd64+arm64) with
   [GoReleaser](.goreleaser.yml) and publishes them as a GitHub Release with
   checksums. **This step alone is enough** for
   `pulumi plugin install resource garage <version>` to work - the provider
   resolves straight from this repo's GitHub Releases
   (`WithPluginDownloadURL` in `provider/provider.go`), no Pulumi Registry
   listing required.
2. Pushes a second, path-prefixed tag
   (`sdk/go/pulumi-garage/vX.Y.Z`) so
   `go get github.com/axnic/pulumi-garage/sdk/go/pulumi-garage@vX.Y.Z`
   resolves to a clean version instead of a pseudo-version (Go's module
   versioning rule for modules living in a subdirectory of a repo - see
   [go.dev/ref/mod#vcs-version](https://go.dev/ref/mod#vcs-version)).
3. Publishes the other 3 SDKs (nodejs, python, dotnet) to their respective
   registries. Each registry's credential is optional (see
   [Registry secrets](#registry-secrets)).

The mechanics of the central workflow (authentication, trusted publishing,
dist-tags, what it gates on) live in the
[`axnic/.github` wiki](https://github.com/axnic/.github/wiki), not here.

There's no Java/Maven SDK - dropped as not worth the setup cost (Sonatype
namespace verification, GPG-signed releases) given how little of the
Pulumi ecosystem uses Java.

A version with a prerelease segment (`1.2.0-rc.1`, ...) is automatically
marked as a prerelease on GitHub (`release.prerelease: auto` in
`.goreleaser.yml`).

## Registry secrets

Configure these as [repository secrets](https://github.com/axnic/pulumi-garage/settings/secrets/actions).
None are required to release the provider binary itself - only to publish
the corresponding language SDK. Each is optional per registry.

| Secret | Registry | Used by |
| --- | --- | --- |
| `NPM_TOKEN` | [npmjs.com](https://www.npmjs.com) | Node.js SDK (`@axnic/pulumi-garage`) |
| `PYPI_API_TOKEN` | [PyPI](https://pypi.org) | Python SDK (`pulumi_garage`) |
| `NUGET_API_KEY` | [NuGet.org](https://www.nuget.org) | .NET SDK (`Axnic.Pulumi.Garage`) |

Any registry-side setup (trusted publishers, token scopes) and the exact
behaviour when a secret is missing are documented in the
[`axnic/.github` wiki](https://github.com/axnic/.github/wiki).

NuGet package naming: the package was originally named `Pulumi.Garage`,
matching the npm/PyPI convention, but that name never actually published
(`dotnet nuget push` reported "already exists" despite the package never
appearing on NuGet). It was renamed to `Axnic.Pulumi.Garage` to sidestep the
question (and any collision with the `Pulumi.*` prefix used by Pulumi's own
official/partner providers).

## Cutting a release

1. Make sure `main` is green (CI passing) and has everything you want to
   release.
2. Dispatch the release, from the Actions tab ("Release" workflow, "Run
   workflow") or the CLI. Provide `bump` **or** `version`, never both:
   ```sh
   gh workflow run workflow_dispatch.release.yaml -f bump=auto
   gh workflow run workflow_dispatch.release.yaml -f version=1.2.0
   # release candidate
   gh workflow run workflow_dispatch.release.yaml -f version=1.2.0-rc.1
   ```
3. Watch the [`Release` workflow
   run](https://github.com/axnic/pulumi-garage/actions/workflows/workflow_dispatch.release.yaml).
   The GitHub Release itself (provider binaries) always runs; SDK publishing
   depends on the registry secrets above.
4. Verify: `pulumi plugin install resource garage <version>` (or let `pulumi
   up` resolve it automatically from the provider's `pluginDownloadURL`),
   and check the registries you published to for the new package version.

## Retrying a failed publish

Publishing is not currently idempotent-safe to blindly re-run for every
registry (npm and PyPI both reject re-publishing the same version; NuGet's
`--skip-duplicate` tolerates retries better). If one SDK's publish step
fails:

- Fix the underlying issue (expired token, missing namespace claim, etc.).
- Re-run just the failed job from the Actions UI ("Re-run failed jobs") —
  the GitHub Release and Go SDK tag steps are idempotent and safe to
  re-trigger; check the registry's own state before retrying npm/PyPI if
  partial upload is a concern.
