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
| README | H2 sections: Prerequisites/Quickstart/Resource reference/Provider config/Dev & testing/Known limitations/Additional details | Terse, no marketing fluff, no emojis | README.md |
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
- nodejs/python/dotnet/java SDK dirs were deleted from sdk/ and examples/ as part of
  the boilerplate->real-provider conversion; don't reintroduce references to them
  without checking they've actually been regenerated.
- CI job list (as of last check): lint, commitlint, build, test, e2e — defined in
  .github/workflows/merge_group,pull_request,push.ci.yaml. Triggers on
  merge_group, pull_request, AND push to main (commits land on main directly,
  without a PR, so the push trigger is what actually validates them).
