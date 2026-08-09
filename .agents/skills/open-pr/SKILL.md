---
name: open-pr
description: >
  Opens a well-formed pull request for pulumi-garage. Use when asked to
  create, open, or submit a pull request, or to push a branch and request a
  review. Enforces correct branch naming, conventional commit title, signed-off
  commits, the project PR template, and a green CI suite before submitting.
compatibility: Requires git and GitHub CLI (gh)
allowed-tools: Bash(git:*) Bash(gh:*) Bash(make:*)
---

# Open Pull Request Skill

## Pre-flight checks

Before pushing anything, read these files to understand conventions:

- `AGENT.md` — architecture, key conventions, prior gotchas
- `CONTRIBUTING.md` — setup, tooling, test layers, commit format
- `.commitlintrc.js` — enforced commit types and scopes (or read the
  `commit` skill, which mirrors it exactly)

There is no `CHANGELOG.md` to update — release notes are generated from
conventional-commit history by goreleaser at tag time (`changelog: use:
github` in `.goreleaser.yml`), not hand-maintained per PR.

Run the same checks CI runs, and confirm they're green:

```sh
make provider
make test
make lint
```

Never open a PR with a failing check suite. Fix the issue first.

## Branch naming

Branch from `main`. Prefix with the commit type the PR's primary change
uses (see the `commit` skill's Types table):

| Primary change | Branch pattern                 |
| ----------------- | ------------------------------- |
| Bug fix            | `fix/<short-description>`      |
| New feature        | `feat/<short-description>`     |
| Documentation      | `docs/<short-description>`     |
| CI/CD              | `ci/<short-description>`       |
| Tooling/build      | `tooling/<short-description>`  |
| Refactor           | `refactor/<short-description>` |

## Commits

Every commit must follow the `commit` skill exactly (type, one mandatory
scope from `provider · sdk · examples · tests · docs · ci · deps · tooling`,
sentence-case subject, WHY-focused body, DCO `Signed-off-by` via `git commit
-s`, `Assisted-by:` trailer — never `Co-authored-by:`). Read that skill
before drafting commits for this PR; don't duplicate its rules here.

Cryptographic signing (`-S`) happens automatically on this machine
(`commit.gpgsign=true`) - `git commit -s` is enough.

## Opening the PR

Push the branch, then fill in `.github/PULL_REQUEST_TEMPLATE.md` (don't
submit it with placeholders still in it) and create the PR:

```sh
git push -u origin <branch-name>
gh pr create \
  --title "<type>(<scope>): <Subject in sentence case>" \
  --body-file /tmp/pr-body.md \
  --base main
```

## Filling the PR template

The template (`.github/PULL_REQUEST_TEMPLATE.md`) has these sections; fill
every one, don't delete the checklist.

### Summary

One sentence, present tense, mirroring the primary commit subject.

### Why

Explain the motivation. `Closes #<number>` if there's a linked issue
(auto-closes on merge), otherwise leave the placeholder as `N/A`.

### What changed

Brief bullet list of the approach and key changes - reviewers should
understand it without reading every diff line.

### How to validate

Already pre-filled with `make provider` / `make test` / `make lint`. Add any
extra manual steps (e.g. `make test_e2e` for a change touching resource
lifecycle behavior) below it.

### Impact

Check "No breaking changes", or check "Breaking change" and describe it
(and use `type(scope)!:` with a `BREAKING CHANGE:` footer on the commit
itself - see the `commit` skill).

### Checklist

Tick every box that's actually true - Code quality (`make test`/`make
lint` pass locally, new behavior covered by tests), Documentation (schema
reflects public provider changes, README updated if the public API
changed, AI-authorship noted if applicable), Commits (Conventional
Commits with a single required scope, each commit focused), Legal (DCO
`Signed-off-by` on every commit, CLA note if contributing under an
employer's copyright). Don't check a box that isn't true.

## Responding to review feedback

For small fixes, amend the commit and force-push:

```sh
git add <files>
git commit --amend --signoff --no-edit
git push --force-with-lease
```

For larger review rounds, prefer a new commit (easier to diff):

```sh
git commit -s -m "fix(provider): Address review: validate bucket ID before create"
```

## CI

`.github/workflows/merge_group,pull_request,push.ci.yaml` runs four required
checks on every PR: Lint, Commit Messages, Build, Tests (the last now
includes a minimum test-coverage gate). Branch protection on `main` requires
all four before merge. If any check is red, fix it in a new commit - do not
skip hooks or force-merge.
