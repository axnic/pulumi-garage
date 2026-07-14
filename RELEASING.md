# Releasing

This document is for maintainers cutting a release of the provider and its SDKs.

## How a release works

Pushing a tag matching `v*` (e.g. `v1.0.0`, or `v1.0.0-alpha.1` for a prerelease)
triggers [`.github/workflows/push.release.yaml`](.github/workflows/push.release.yaml),
which:

1. Builds the provider binary for darwin/linux/windows (amd64+arm64) with
   [GoReleaser](.goreleaser.yml) and publishes them as a GitHub Release with
   checksums. **This step alone is enough** for
   `pulumi plugin install resource garage <version>` to work — the provider
   resolves straight from this repo's GitHub Releases
   (`WithPluginDownloadURL` in `provider/provider.go`), no Pulumi Registry
   listing required.
2. Pushes a second, path-prefixed tag
   (`sdk/go/pulumi-garage/vX.Y.Z`) so
   `go get github.com/axnic/pulumi-garage/sdk/go/pulumi-garage@vX.Y.Z`
   resolves to a clean version instead of a pseudo-version (Go's module
   versioning rule for modules living in a subdirectory of a repo — see
   [go.dev/ref/mod#vcs-version](https://go.dev/ref/mod#vcs-version)).
3. Publishes the other 4 SDKs (nodejs, python, dotnet, java) to their
   respective registries. **Each is independently gated on its own secret(s)
   being configured** — a registry you haven't onboarded yet is silently
   skipped, not failed, so you can cut binary-only releases before any
   registry is set up, and onboard registries one at a time afterwards.

A tag with a prerelease segment (`-alpha.1`, `-beta.2`, ...) is automatically
marked as a prerelease on GitHub (`release.prerelease: auto` in
`.goreleaser.yml`) and published to npm under the `next` dist-tag instead of
`latest` (npm refuses to publish a prerelease version to `latest`).

## Required secrets

Configure these as [repository secrets](https://github.com/axnic/pulumi-garage/settings/secrets/actions).
None are required to release the provider binary itself — only to publish
the corresponding language SDK.

| Secret | Registry | Used by |
| --- | --- | --- |
| `NPM_TOKEN` | [npmjs.com](https://www.npmjs.com) | Node.js SDK (`@axnic/pulumi-garage`) |
| `PYPI_API_TOKEN` | [PyPI](https://pypi.org) | Python SDK (`pulumi_garage`) |
| `NUGET_API_KEY` | [NuGet.org](https://www.nuget.org) | .NET SDK (`Pulumi.Garage`) |
| `PUBLISH_REPO_USERNAME`, `PUBLISH_REPO_PASSWORD` | [Maven Central / Sonatype](https://central.sonatype.com) | Java SDK (`com.axnic.pulumi:pulumi-garage`) |
| `SIGNING_KEY`, `SIGNING_PASSWORD` | GPG signing | Java SDK — Maven Central rejects unsigned artifacts |

### Obtaining each secret

- **`NPM_TOKEN`** — create an npm [Automation
  token](https://docs.npmjs.com/creating-and-viewing-access-tokens) scoped to
  the `@axnic` org, with publish access.
- **`PYPI_API_TOKEN`** — create a [PyPI API
  token](https://pypi.org/help/#apitoken) scoped to the `pulumi_garage`
  project (or your account, before the project exists yet).
- **`NUGET_API_KEY`** — create a [NuGet API
  key](https://learn.microsoft.com/en-us/nuget/nuget-org/publish-a-package#create-api-keys)
  scoped to the `Pulumi.Garage` package.
- **`PUBLISH_REPO_USERNAME` / `PUBLISH_REPO_PASSWORD`** — register a
  [Sonatype Central account](https://central.sonatype.com/) and claim the
  `com.axnic.pulumi` namespace (requires proving control of the `axnic`
  GitHub org or a domain you own), then generate a [user
  token](https://central.sonatype.org/publish/generate-portal-token/) — the
  token's username/password pair are these two secrets.
- **`SIGNING_KEY` / `SIGNING_PASSWORD`** — generate a dedicated GPG key pair
  for releases (`gpg --full-generate-key`), publish the public key to a
  keyserver (Maven Central verifies signatures against
  `keys.openpgp.org`/`keyserver.ubuntu.com`), then export the private key as
  an ASCII-armored, in-memory-usable key:
  ```sh
  gpg --export-secret-keys --armor <key-id> > signing-key.asc
  ```
  `SIGNING_KEY` is the contents of `signing-key.asc`; `SIGNING_PASSWORD` is
  the key's passphrase. The Gradle build consumes both via
  `useInMemoryPgpKeys` (see `sdk/java/build.gradle`) — no keyring file needed
  on the runner.

## Cutting a release

1. Make sure `main` is green (CI passing) and has everything you want to
   release.
2. Tag and push:
   ```sh
   git tag v1.0.0
   git push origin v1.0.0
   ```
   For a prerelease: `git tag v1.0.0-alpha.1`.
3. Watch the [`Release` workflow
   run](https://github.com/axnic/pulumi-garage/actions/workflows/push.release.yaml).
   The GitHub Release itself (provider binaries) always runs; the SDK
   publish steps run only for the registries you've configured secrets for.
4. Verify: `pulumi plugin install resource garage <version>` (or let `pulumi
   up` resolve it automatically from the provider's `pluginDownloadURL`),
   and check the registries you published to for the new package version.

## Retrying a failed publish

Publishing is not currently idempotent-safe to blindly re-run for every
registry (npm and PyPI both reject re-publishing the same version; NuGet's
`--skip-duplicate` and Maven's staging flow tolerate retries better). If one
SDK's publish step fails:

- Fix the underlying issue (expired token, missing namespace claim, etc.).
- Re-run just the failed job from the Actions UI ("Re-run failed jobs") —
  the GitHub Release and Go SDK tag steps are idempotent and safe to
  re-trigger; check the registry's own state before retrying npm/PyPI if
  partial upload is a concern.
