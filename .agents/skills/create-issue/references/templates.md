# Issue Template Field Reference

Source: `.github/ISSUE_TEMPLATE/bug.yaml`, `feature.yml`, and `config.yml`.

## Bug report (`bug.yaml`)

No Area or Package-version dropdown — `pulumi about`'s output covers
version info, and the report is free-form rather than categorized.

| Field (form ID)          | Required | Notes                                                              |
| -------------------------- | -------- | --------------------------------------------------------------------- |
| Describe what happened (`what-happened`) | Yes | Include Pulumi commands run and an inline snippet of any error/output |
| Sample program (`sample-program`)        | Yes | A minimal, self-contained Pulumi program reproducing the behavior     |
| Log output (`log-output`)                | No  | Output of `pulumi up --logtostderr --logflow -v=10`                  |
| Affected Resource(s) (`resources`)       | No  | e.g. `garage:Bucket`, `garage:Key`, `garage:BucketKeyPermission`      |
| Output of `pulumi about` (`versions`)    | Yes | Full output of `pulumi about` from the project root                  |
| Additional context (`ctx`)               | No  | Anything else worth adding                                            |

Labels applied: `kind/bug`, `needs-triage`.

## Feature request (`feature.yml`)

### Area options

```
- Provider – resource or datasource logic
- SDK – generated SDK (Node.js / Python / Go / .NET)
- Examples – example programs
- Tests – provider or SDK tests
- Docs / README
- CI / CD
- Tooling / build
- Not sure
```

### Type of change options

```
- New capability (adds something that doesn't exist)
- Improvement (enhances an existing feature)
- Developer experience (DX / ergonomics)
- Performance
- Other
```

### Required vs optional fields

| Field (form ID)                | Required |
| --------------------------------- | -------- |
| Summary (`summary`)              | Yes      |
| Area (`area`)                    | Yes      |
| Type of change (`type`)          | Yes      |
| Motivation / problem (`motivation`) | Yes   |
| Proposed solution (`proposal`)   | Yes      |
| Alternatives considered (`alternatives`) | No |
| Created by AI (`ai_generated`, checkbox) | No — check it when this issue was filed by an AI agent |
| Additional context (`extra`)     | No       |

Labels applied: `enhancement`, `needs-triage`.

## Repo contact links (from `config.yml`)

`blank_issues_enabled: false` — every issue must go through one of the two
templates above, or one of these contact links:

- **Security vulnerability** → <https://github.com/axnic/pulumi-garage/security/advisories/new>
  _(Do NOT open a public issue — always redirect the user here)_
- **Question / discussion** → <https://github.com/axnic/pulumi-garage/discussions>
  _(For questions and ideas that aren't actionable bugs or features, redirect here)_
