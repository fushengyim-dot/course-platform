## 1. Review contract

- [x] 1.1 Add the seven code categories and seven scoring dimensions to the existing review entrypoints, with links to current authoritative standards and specs.
- [x] 1.2 Add a review record template that distinguishes self-review, automated finding and independent peer review.
- [x] 1.3 Add a CI check requiring a well-formed `docs/reviews/` record on every PR targeting `main`.

## 2. Evidence and correction

- [x] 2.1 Review the current synthetic public-information handler, search service and fixture adapter; create three evidence-based records with accurate scope and conclusions.
- [ ] 2.2 Correct the confirmed 503 response-code mismatch and add tests for repository failure responses.
- [ ] 2.3 Run strict OpenSpec validation, the project verification workflow, the review-record checker, secret/history scan and diff review on the same revision.
- [x] 2.4 Record limitations, including the absence of independent peer review; do not represent a single-maintainer self-review as cross-review.
