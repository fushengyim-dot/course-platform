# Design

## Scope and source

This change implements the course requirement to inspect code by seven categories and record findings against seven review dimensions. The checklist asks: what the code intends to guarantee, which mechanism enforces it, when that mechanism can fail, and how to correct it. The second layer rates requirement coverage, consistency, failure paths, query performance, duplicated logic, vulnerabilities and high-risk behavior. Findings use the project's three conclusions: usable, needs modification, or ineffective/high risk.

The checklist lives under `docs/code-review/` and is linked from the repository entry rules and the existing review standard. It does not replace product, backend, security, testing, operations, AI, data or review standards. Each checklist item points to the relevant OpenSpec scenario or standard where one exists; missing evidence is recorded as unknown or not applicable, never assumed green.

## Review record and merge gate

Each pull request targeting `main` includes at least one Markdown record under `docs/reviews/`. The record identifies the reviewed revision and files, review method and reviewer role, applicable requirements, seven-category inspection, seven-dimension rating, findings with evidence, remediation or follow-up, and unresolved limits. A record may document self-review, automated discovery, or cross-review; these roles must be labelled accurately. CI checks that a record is included and has the required fields, while a human reads the technical reasoning. The script does not claim to prove review quality or reviewer independence.

When the record is added after the code has been reviewed, it names the exact code revision that was inspected. Any subsequent code change invalidates the prior review for the affected lines and requires re-review. The required CI check runs on the PR base/head pair; it does not use a stale branch-wide report as evidence.

The repository currently has one maintainer. This change enables self-review and scripted discovery records but does not claim that a second-person review has occurred. If the course requires independent peer review, that evidence remains outstanding until a real peer performs it.

## Confirmed finding in the current S1 code

The existing handler returns HTTP 503 with `UNAVAILABLE` for several public-information and institution repository failures. The registered backend contract and the active published-information design require `DEPENDENCY_UNAVAILABLE`. Correct those responses and add handler-level failure-path assertions. Keep public messages generic and do not expose repository errors.

## Explicit non-goals

- No fake defect or intentionally vulnerable production code is introduced to create a red finding.
- No database is added and no claim is made about SQL query plans or indexes while the public dataset is an in-memory synthetic fixture.
- No real identity, personal data, source ingestion, AI provider, message queue, or public deployment is enabled.

## Rollback

Revert this single PR. The new check is additive and has no data migration or runtime configuration. If the record checker fails incorrectly, fix or revert the checker in a follow-up PR; do not bypass branch protection.
