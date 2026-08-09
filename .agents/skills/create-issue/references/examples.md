# Issue Examples

Good and bad examples to calibrate quality before creating an issue.

## Title quality

### Bug report titles

| ✅ Good                                                       | ❌ Bad                     |
| ---------------------------------------------------------------- | --------------------------- |
| `Bucket delete fails silently when the bucket still has objects` | `Bucket delete bug`         |
| `Key.createBucket permission is dropped after an Update`         | `Permission issue`          |
| `Provider panics when GARAGE_ADMIN_TOKEN is set but empty`       | `Fix crash`                 |
| `BucketKeyPermission read doesn't detect a revoked grant`        | `Permission read doesn't work` |

**Rule:** The title should describe the **symptom** or **broken behaviour** —
not the suspected fix, and not a generic label like "bug" or "issue".

### Feature request titles

| ✅ Good                                                        | ❌ Bad                    |
| ------------------------------------------------------------------ | --------------------------- |
| `Add support for Garage bucket lifecycle configuration`            | `Lifecycle support`         |
| `Support local (per-key) bucket aliases in Bucket`                 | `More alias support`        |
| `Model Key's global createBucket permission`                       | `createBucket flag`         |
| `Generate a Java/Maven SDK`                                        | `Java support`              |

**Rule:** The title should describe **what capability is added** or **what
changes** — concrete enough that a maintainer understands scope immediately.

---

## Completed issue bodies

### Bug report example

**Title:** `Bucket delete fails silently when the bucket still has objects`

```markdown
### Describe what happened

Running `pulumi destroy` on a stack with a non-empty `garage:Bucket`
reports success, but the bucket and its objects are still present when
checked against the Garage Admin API afterwards.

### Sample program

\`\`\`yaml
name: repro
runtime: yaml
resources:
  myBucket:
    type: garage:Bucket
    properties:
      globalAlias: repro-bucket
\`\`\`
(then put at least one object in the bucket via the S3 API before destroying)

### Log output

(paste of `pulumi up --logtostderr --logflow -v=10` output, or a Gist link
if long)

### Affected Resource(s)

garage:Bucket

### Output of `pulumi about`

(paste of `pulumi about`)

### Additional context

Expected Garage's own `BucketNotEmpty` error to propagate and fail the
destroy instead.
```

---

### Feature request example

**Title:** `Add support for Garage bucket lifecycle configuration`

````markdown
### Summary

Add a `lifecycleRules` block to the `Bucket` resource so objects can be
expired automatically via Garage's lifecycle API, instead of requiring a
separate script.

### Area

Provider – resource or datasource logic

### Type of change

New capability (adds something that doesn't exist)

### Motivation / problem

Garage supports per-bucket lifecycle rules (expiration, abort incomplete
multipart uploads) via its Admin API, but this provider has no way to
configure them - users currently have to call the Garage CLI or Admin API
directly outside of Pulumi, which the rest of the bucket's config doesn't
need.

### Proposed solution

```yaml
resources:
  myBucket:
    type: garage:Bucket
    properties:
      lifecycleRules:
        - id: expire-old-uploads
          prefix: tmp/
          expiration:
            days: 7
```

Maps onto Garage's `PUT /v2/UpdateBucket` `lifecycle` field the same way
`website`/`quotas` are already handled in `bucket_resource.go`.

### Alternatives considered

A separate `BucketLifecycleRule` resource (mirroring `BucketKeyPermission`)
was considered, but Garage's lifecycle config is a single ordered list per
bucket, not independently addressable grants - a nested block matches the
underlying API shape better.

### Additional context

See Garage's [lifecycle
documentation](https://garagehq.deuxfleurs.fr/documentation/reference-manual/lifecycle-configuration/).
````
