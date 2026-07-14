# Pulumi Garage Provider

A Pulumi native provider for managing [Garage](https://garagehq.deuxfleurs.fr), a
self-hosted, S3-compatible distributed object storage system, through its Admin API.

It manages three resources:

- **`Bucket`** - a Garage bucket, with an optional global alias, static-website
  hosting configuration, and storage quotas.
- **`Key`** - an S3 access key.
- **`BucketKeyPermission`** - a read/write/owner permission grant of a `Key` on a
  `Bucket`.

Cluster bootstrapping (nodes, zones, capacity, layout) is out of scope - see
[Known limitations](#known-limitations).

> **Note on provenance.** This provider was built end-to-end by an AI coding
> agent (Claude) in an autonomous, "vibe coded" session, without a human
> reviewing the implementation line by line. That's a deliberate reason the
> project leans hard on automated testing rather than manual review for
> confidence: unit tests for the Admin API client and each resource's
> create/read/update/delete logic, plus a real end-to-end suite that stands up
> an actual Garage cluster in Docker and exercises it (including a genuine S3
> object upload/download through a granted permission) - see
> [Development & testing](#development--testing). Review the code yourself
> before trusting it with production data.

## Prerequisites

- [Go](https://go.dev/dl/) 1.25 or later (to build the provider binary)
- [Pulumi CLI](https://www.pulumi.com/docs/iac/download-install/)
- [Docker](https://docs.docker.com/get-docker/) - only needed to run the local
  end-to-end tests (`make test_e2e`), not to use the provider itself
- A running Garage cluster, with its Admin API reachable. This provider does not
  bootstrap one for you; see [garagehq.deuxfleurs.fr](https://garagehq.deuxfleurs.fr)
  for setting one up (`garage layout assign`/`apply`, or `--single-node` for a
  single-node cluster).

## Quickstart

Build and install the provider plugin:

```
make build install
```

Point it at your Garage cluster's Admin API, either via stack config:

```
pulumi config set garage:endpoint http://localhost:3903
pulumi config set garage:adminToken <token> --secret
```

or via the `GARAGE_ADMIN_ENDPOINT` / `GARAGE_ADMIN_TOKEN` environment variables.

Then, a minimal YAML program (see `examples/yaml/Pulumi.yaml`):

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

Only YAML and Go example programs exist and are tested for this v1; nodejs,
python, dotnet, and java SDKs are not generated (see
[Known limitations](#known-limitations)). The Go equivalent is at
`examples/go/main.go`.

## Resource reference

### `Bucket` (`garage:index:Bucket`)

Manages a Garage bucket: its (single) global alias, static-website hosting
configuration, and storage quotas.

Inputs:
- `globalAlias` (optional) - The bucket's human-readable global alias, e.g.
  `"my-app-data"`. Buckets can be created without one and addressed only by ID,
  but an alias is required to use the bucket over the S3 API with most clients.
  Only a single global alias is supported; local (per-key) aliases and multiple
  global aliases are not managed by this provider.
- `website` (optional) - Static-website hosting configuration for the bucket.
  Omit to leave website hosting disabled.
  - `indexDocument` - The document served for requests to the bucket root or any
    "directory".
  - `errorDocument` (optional) - The document served for requests that don't
    match an existing object. Defaults to Garage's built-in error page if unset.
- `quotas` (optional) - Storage quotas for the bucket. Omit either field, or the
  whole block, to leave that limit unset.
  - `maxSize` (optional) - The maximum total size, in bytes, the bucket may hold.
    Unset means no limit.
  - `maxObjects` (optional) - The maximum number of objects the bucket may hold.
    Unset means no limit.

Outputs (in addition to the inputs above):
- `createdAt` - The RFC 3339 timestamp at which the bucket was created.
- `objects` - The number of objects currently stored in the bucket.
- `bytes` - The total size, in bytes, of all objects currently stored in the
  bucket.

Deleting a non-empty bucket fails, mirroring Garage's (and S3's) own semantics -
empty it first before removing it from your Pulumi program.

### `Key` (`garage:index:Key`)

Manages a Garage S3 access key.

Inputs:
- `name` (optional) - A human-readable label for the key. If unset, Garage
  assigns a default name.

Outputs:
- `accessKeyId` - The S3 access key ID, e.g. the value of `AWS_ACCESS_KEY_ID`.
- `secretAccessKey` - The S3 secret access key, e.g. the value of
  `AWS_SECRET_ACCESS_KEY`. Only ever readable at creation time - Garage's Admin
  API does not return it again afterwards, so it is captured once at `Create`
  and carried forward in state.
- `createdAt` - The RFC 3339 timestamp at which the key was created.

The key's global `createBucket` permission is not modelled in v1; keys are
scoped to buckets exclusively via `BucketKeyPermission`.

### `BucketKeyPermission` (`garage:index:BucketKeyPermission`)

Grants an access `Key` read/write/owner permissions on a `Bucket`. Garage has no
single natural ID for this grant, so the resource ID is a synthetic
`<bucketId>/<accessKeyId>` composite.

Inputs:
- `bucketId` - The ID of the `Bucket` to grant permissions on.
- `accessKeyId` - The access key ID of the `Key` to grant permissions to.
- `permissions` - The read/write/owner permissions to grant.
  - `read` (optional) - Whether the key can read objects (`GetObject`,
    `ListObjects`, ...) from the bucket.
  - `write` (optional) - Whether the key can write objects (`PutObject`,
    `DeleteObject`, ...) to the bucket.
  - `owner` (optional) - Whether the key has owner rights on the bucket (manage
    bucket-level settings such as its website configuration or quotas via the
    S3 API).

## Provider configuration reference

| Key | Env var fallback | Description |
|---|---|---|
| `garage:endpoint` | `GARAGE_ADMIN_ENDPOINT` | The base URL of the Garage Admin API, e.g. `"http://localhost:3903"`. |
| `garage:adminToken` (secret) | `GARAGE_ADMIN_TOKEN` | A bearer token authorized against the Garage Admin API. |

Both are required, one way or the other - the provider fails to configure if
neither the config key nor its env var fallback is set.

## Development & testing

Build and install the provider binary:

```
make build install
```

After changing resource code, regenerate the schema (`provider/cmd/pulumi-resource-garage/schema.json`)
and Go SDK:

```
make codegen
```

Tests are layered:

- `make test` - fast, hermetic unit tests. No live Garage cluster required;
  example-lifecycle tests are skipped (via `t.Skip`) when
  `GARAGE_ADMIN_ENDPOINT` isn't set.
- `make test_e2e` - spins up a real, disposable, single-node Garage cluster via
  Docker Compose (`docker-compose.e2e.yml`, `dxflrs/garage:v2.3.0`), runs the
  example programs' full create/update/delete lifecycle against it plus a real
  S3 `PutObject`/`GetObject` round trip, then tears it down. Requires Docker.
- `make lint` - runs `golangci-lint`.

CI (`.github/workflows/merge_group,pull_request,push.ci.yaml`) runs lint,
commitlint, build, unit tests, and the full E2E suite on every pull request,
merge-queue entry, and push to `main` - this repo takes commits directly on
`main` without a PR, so the push trigger is what actually validates them.

The provider's Admin API client (`provider/internal/garageclient/`) is a
hand-written, thin HTTP client, internal to the provider and not part of its
public surface.

## Known limitations

This is a v1, personal/small-org provider - scope is intentionally narrow:

- **No cluster layout management.** Assigning nodes, zones, and capacity is not
  managed by this provider; it's a one-shot bootstrap operation that doesn't fit
  cleanly into idempotent IaC. Bootstrap your Garage cluster yourself (e.g.
  `garage layout assign`/`apply`, or `--single-node`) before pointing this
  provider at it.
- **`Bucket` supports only a single global alias.** No local (per-key) aliases,
  no multiple global aliases.
- **`Key`'s global `createBucket` permission is not modelled**, deliberately -
  it wasn't reliable enough to model without further verification against a
  live API.
- **Deleting a non-empty `Bucket` fails.** This mirrors Garage's (and S3's) own
  native behavior; it's by design, not a bug.
- **Only YAML and Go example programs exist and are tested.** nodejs, python,
  dotnet, and java SDKs are not generated for this v1 - their stale boilerplate
  code was removed rather than fixed.

## Additional details

This provider is built on [`pulumi-go-provider`](https://github.com/pulumi/pulumi-go-provider);
see its docs for background on how native Pulumi providers work.

Repo layout:
- `provider/` - the provider implementation (`provider.go`, `config.go`,
  `*_resource.go`), plus the internal Garage Admin API client
  (`provider/internal/garageclient/`) and the codegen entrypoint
  (`provider/cmd/pulumi-resource-garage/`).
- `sdk/` - the generated Go SDK (`make codegen`).
- `examples/` - the YAML and Go example programs, and their lifecycle tests.
- `docker-compose.e2e.yml`, `test/e2e/garage.toml` - the local/CI E2E fixture:
  a disposable single-node Garage cluster with a fixed, test-only
  `rpc_secret`/`admin_token` (not sensitive - the cluster is ephemeral and
  local-only).
- `Makefile` - build, codegen, lint, and test targets.
