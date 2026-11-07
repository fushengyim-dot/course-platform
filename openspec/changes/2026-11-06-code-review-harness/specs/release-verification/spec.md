## ADDED Requirements

### Requirement: Pull request review evidence
Every pull request targeting `main` SHALL include a review record under `docs/reviews/` that identifies the reviewed code revision, scope, review method, applicable requirements, findings, conclusion and unresolved limits. The record SHALL distinguish self-review, automated discovery and independent peer review. A machine check SHALL verify that the record is present and structurally complete; a maintainer SHALL review its reasoning before merge.

#### Scenario: Pull request includes review evidence
- **WHEN** a pull request targeting `main` changes the proposed revision
- **THEN** CI verifies that at least one review record is added or updated in that pull request and has the required fields

#### Scenario: Missing or incomplete review record
- **WHEN** a pull request has no review record or the record omits required scope, revision, method, criteria, conclusion or limitations
- **THEN** the required verification check fails and the pull request is not eligible for merge

#### Scenario: Code changes after review
- **GIVEN** the reviewed code changes after the record was written
- **WHEN** the pull request head is updated
- **THEN** the code is reviewed again and the record identifies the new inspected revision

#### Scenario: Single-maintainer review
- **GIVEN** only one maintainer performs the review
- **WHEN** the record is checked
- **THEN** it is labelled self-review or automated discovery and does not claim independent peer review
