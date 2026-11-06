# Why

The project has eight engineering standards, but course code review requires a separate seven-category inspection method, scored review records and a repeatable merge gate. The current PR workflow runs tests and scans, but does not require a review record that explains what code was checked and why the conclusion is justified.

# What Changes

- Add a seven-category code inspection checklist and a seven-dimension review record format to the existing review workflow.
- Require a review record for every pull request targeting `main`, and validate the record's basic structure in CI.
- Record focused reviews of the existing synthetic public-information implementation and correct any confirmed contract mismatch found during review.
- Keep the existing eight engineering standards authoritative for their current topics; the new checklist is a code-reading and evidence-capture layer, not a parallel set of engineering rules.

# Impact

This changes repository development and review evidence only. It adds no business capability, production identity, real student data, real admissions source, model provider, external service or deployment.
