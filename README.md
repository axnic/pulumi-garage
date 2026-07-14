[![GitHub release](https://img.shields.io/github/v/release/axnic/pulumi-garage?logo=github&sort=semver)](https://github.com/axnic/pulumi-garage/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/axnic/pulumi-garage/sdk/go/pulumi-garage.svg)](https://pkg.go.dev/github.com/axnic/pulumi-garage/sdk/go/pulumi-garage)
[![npm version](https://img.shields.io/npm/v/%40axnic%2Fpulumi-garage.svg)](https://www.npmjs.com/package/@axnic/pulumi-garage)
[![PyPI version](https://img.shields.io/pypi/v/pulumi_garage.svg)](https://pypi.org/project/pulumi_garage/)
[![NuGet version](https://img.shields.io/nuget/v/Axnic.Pulumi.Garage.svg)](https://www.nuget.org/packages/Axnic.Pulumi.Garage/)
[![License](https://img.shields.io/github/license/axnic/pulumi-garage.svg)](LICENSE)

# Pulumi Garage Provider

A Pulumi native provider for managing [Garage](https://garagehq.deuxfleurs.fr), a
self-hosted, S3-compatible distributed object storage system, through its Admin API.

It manages three resources: **`Bucket`** (with an optional global alias, static-website
hosting, and storage quotas), **`Key`** (an S3 access key), and
**`BucketKeyPermission`** (a read/write/owner grant of a `Key` on a `Bucket`). Cluster
bootstrapping (nodes, zones, capacity, layout) is out of scope - see
[Known limitations](#known-limitations).

> [!NOTE]
> **Provenance.** This provider was built end-to-end by an AI coding agent (Claude)
> in an autonomous, "vibe coded" session, without a human reviewing the implementation
> line by line. That's a deliberate reason the project leans hard on automated testing
> rather than manual review for confidence: unit tests for the Admin API client and
> each resource's create/read/update/delete logic, plus a real end-to-end suite that
> stands up an actual Garage cluster in Docker and exercises it (including a genuine S3
> object upload/download through a granted permission) - see [CONTRIBUTING.md](CONTRIBUTING.md).
> Review the code yourself before trusting it with production data.

## Installing

This package is available in several languages. Only the Go and YAML SDKs are
exercised by this repo's own examples and E2E tests (see
[Known limitations](#known-limitations)); the others are generated and published for
convenience but are otherwise untested by this project beyond schema-level checks.

### Node.js (JavaScript/TypeScript)

```bash
npm install @axnic/pulumi-garage
```

### Python

```bash
pip install pulumi_garage
```

### Go

```bash
go get github.com/axnic/pulumi-garage/sdk/go/pulumi-garage
```

### .NET

```bash
dotnet add package Axnic.Pulumi.Garage
```

The provider binary itself doesn't require any of the above - `pulumi plugin install
resource garage <version>` (or the engine's automatic resolution) fetches it directly
from this repo's [GitHub Releases](https://github.com/axnic/pulumi-garage/releases), no
Pulumi Registry listing required.

## Quickstart

Point the provider at your Garage cluster's Admin API, either via stack config:

```bash
pulumi config set garage:endpoint http://localhost:3903
pulumi config set garage:adminToken <token> --secret
```

or the `GARAGE_ADMIN_ENDPOINT` / `GARAGE_ADMIN_TOKEN` environment variables. This
provider does not bootstrap a Garage cluster for you - see
[garagehq.deuxfleurs.fr](https://garagehq.deuxfleurs.fr) for setting one up
(`garage layout assign`/`apply`, or `--single-node` for a single-node cluster).

Then, a minimal YAML program (see [`examples/yaml/Pulumi.yaml`](examples/yaml/Pulumi.yaml),
or [`examples/go/main.go`](examples/go/main.go) for the Go equivalent):

```yaml
name: garage-example-yaml
runtime: yaml

resources:
  myKey:
    type: garage:Key
    properties:
      name: my-app-key
  myBucket:
    type: garage:Bucket
    properties:
      globalAlias: my-app-bucket
  myPermission:
    type: garage:BucketKeyPermission
    properties:
      bucketId: ${myBucket.id}
      accessKeyId: ${myKey.accessKeyId}
      permissions:
        read: true
        write: true

outputs:
  bucketId: ${myBucket.id}
  accessKeyId: ${myKey.accessKeyId}
  secretAccessKey: ${myKey.secretAccessKey}
```

## Configuration

| Key | Env var fallback | Description |
|---|---|---|
| `garage:endpoint` | `GARAGE_ADMIN_ENDPOINT` | The base URL of the Garage Admin API, e.g. `"http://localhost:3903"`. |
| `garage:adminToken` (secret) | `GARAGE_ADMIN_TOKEN` | A bearer token authorized against the Garage Admin API. |

Both are required, one way or the other - the provider fails to configure if neither
the config key nor its env var fallback is set.

## Compatibility

This provider talks to Garage's Admin API v2, introduced in Garage v2.0.0. Each
version below is tested against the full example-program lifecycle (create/update/delete
plus a real S3 object upload/download) in its own CI workflow, so its badge reflects
that version alone rather than an aggregate:

| Garage version | Status |
|---|---|
| v2.0.0 | [![E2E (Garage v2.0.0)](https://github.com/axnic/pulumi-garage/actions/workflows/merge_group%2Cpull_request%2Cpush.e2e-garage-2.0.yaml/badge.svg?branch=main)](https://github.com/axnic/pulumi-garage/actions/workflows/merge_group%2Cpull_request%2Cpush.e2e-garage-2.0.yaml) |
| v2.1.0 | [![E2E (Garage v2.1.0)](https://github.com/axnic/pulumi-garage/actions/workflows/merge_group%2Cpull_request%2Cpush.e2e-garage-2.1.yaml/badge.svg?branch=main)](https://github.com/axnic/pulumi-garage/actions/workflows/merge_group%2Cpull_request%2Cpush.e2e-garage-2.1.yaml) |
| v2.2.0 | [![E2E (Garage v2.2.0)](https://github.com/axnic/pulumi-garage/actions/workflows/merge_group%2Cpull_request%2Cpush.e2e-garage-2.2.yaml/badge.svg?branch=main)](https://github.com/axnic/pulumi-garage/actions/workflows/merge_group%2Cpull_request%2Cpush.e2e-garage-2.2.yaml) |
| v2.3.0 | [![E2E (Garage v2.3.0)](https://github.com/axnic/pulumi-garage/actions/workflows/merge_group%2Cpull_request%2Cpush.e2e-garage-2.3.yaml/badge.svg?branch=main)](https://github.com/axnic/pulumi-garage/actions/workflows/merge_group%2Cpull_request%2Cpush.e2e-garage-2.3.yaml) |

Each workflow (`.github/workflows/merge_group,pull_request,push.e2e-garage-2.*.yaml`) is
a thin wrapper calling the reusable `_reusable-e2e.yaml` job with that version pinned -
GitHub Actions status badges are per-workflow-file, not per-matrix-leg, which is why
this is four small files rather than one `strategy.matrix` job. `--single-node` (the
fast layout-bootstrap path) only exists from Garage v2.3.0 onward, so
`docker-compose.yml` and `scripts/bootstrap-garage.sh` always use the manual
`layout assign`/`apply` bootstrap instead, which works identically across all four
versions.

## Known limitations

This is a v1, personal/small-org provider - scope is intentionally narrow:

- **No cluster layout management.** Assigning nodes, zones, and capacity is not
  managed by this provider; it's a one-shot bootstrap operation that doesn't fit
  cleanly into idempotent IaC. Bootstrap your Garage cluster yourself (e.g.
  `garage layout assign`/`apply`, or `--single-node` on Garage v2.3.0+) before
  pointing this provider at it.
- **`Bucket` supports only a single global alias.** No local (per-key) aliases, no
  multiple global aliases.
- **`Key`'s global `createBucket` permission is not modelled**, deliberately - it
  wasn't reliable enough to model without further verification against a live API.
- **Deleting a non-empty `Bucket` fails.** This mirrors Garage's (and S3's) own native
  behavior; it's by design, not a bug.
- **Only YAML and Go example programs are exercised in CI/E2E.** The nodejs, python,
  and dotnet SDKs are generated and published (see [Installing](#installing)) but
  aren't covered by this repo's example programs or lifecycle tests.
- **No Java/Maven SDK.** Pulumi's Java support has too little adoption to justify the
  extra setup cost (Sonatype namespace verification, GPG-signed releases); the nodejs,
  python, dotnet, and go SDKs cover the ecosystem's actual usage.

## Reference

This provider isn't (yet) listed on the [Pulumi Registry](https://www.pulumi.com/registry/),
so there's no `pulumi.com/registry/packages/garage` page to link to for generated,
per-resource API docs. Until then:

- The **Go SDK reference** on [pkg.go.dev](https://pkg.go.dev/github.com/axnic/pulumi-garage/sdk/go/pulumi-garage)
  is generated from the same schema descriptions the Pulumi Registry would render, and
  is the most complete field-by-field reference available (populated after the first
  tagged release - see [RELEASING.md](RELEASING.md)).
- The [`Quickstart`](#quickstart) and [`Configuration`](#configuration) sections above
  cover the three resources and provider config end to end; the underlying schema is
  `provider/cmd/pulumi-resource-garage/schema.json` (generated, not hand-maintained).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for building the provider, the test layers
(including how to run the E2E suite locally), generating the SDKs, and the devcontainer
setup. See also our [Code of Conduct](CODE-OF-CONDUCT.md).

## License

[Apache-2.0](LICENSE)
