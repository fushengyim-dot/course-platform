## 1. Contract approval and data boundary
- [ ] 1.1 Confirm category values, source URL requirement, school/college/program/direction scope fields, aliases and maximum lengths with the product owner.
- [ ] 1.2 Freeze fuzzy-match strategy using representative examples; define evaluation set and evidence quality review without claiming exhaustive coverage.
- [ ] 1.3 Freeze source record, snapshot retention, permissions, crawl frequency/site policy, date semantics, sort/cursor and version contracts.
- [ ] 1.4 Select the server-side identity source and complete the single privacy notice/lifecycle rules before real private data is introduced.

## 2. Source intake and teacher curation
- [ ] 2.1 Implement collection adapters only after their permitted source list and operating policy are approved.
- [ ] 2.2 Store provenance snapshots, content fingerprints, retrieval times and excerpt locators; test restricted and failed sources without bypassing controls.
- [ ] 2.3 Extract multiple candidate items while preserving source lineage; surface duplicates, stale items and conflicts for review.
- [ ] 2.4 Implement teacher edit/reject/approve, immutable revisions, explicit publish/withdraw and atomic public pointer changes.
- [ ] 2.5 Test stale revisions, idempotent replay, failed refresh preserving the public revision and cross-role denial.

## 3. Public and teacher search
- [ ] 3.1 Implement anonymous public home/list/detail and faculty-level fuzzy search over approved public names, aliases and information text.
- [ ] 3.2 Keep student/public response fields separate from teacher evidence candidates and internal source metadata.
- [ ] 3.3 Implement teacher cross-institution direct and cross-disciplinary candidate search with rationale, provenance and review/conflict state.
- [ ] 3.4 Implement typed inclusive date filters, deterministic relevance/date sort and cursor pagination.
- [ ] 3.5 Test anonymous success, unpublished invisibility, invalid queries, null dates, empty results, public storage failure, source link failure and non-exhaustive-result wording.

## 4. Follow-on UI and verification
- [ ] 4.1 Implement the student positioning language guide as a product component/config only after separate scope approval; ground claims in submitted evidence.
- [ ] 4.2 Implement consultation request capture with account/contact provenance; phone and WeChat ID remain separate; no implicit outbound messaging.
- [ ] 4.3 Implement one discoverable privacy notice and teacher access to submitted student files without per-submission approval UI.
- [ ] 4.4 Finalize simulation screens only after the reference flow is available and reviewed.
- [ ] 4.5 Run applicable project verification, browser checks and public/internal boundary review on the same revision before archiving.

## Initial S1 implementation slice (does not close the remaining capability work)
- [x] Owner selected the existing web client, synthetic records, and one-character typo tolerance calibrated with synthetic examples. R26 supersedes the initial student filter set/date choice below.
- [x] Add the anonymous information page at `/information` while preserving the two status routes.
- [x] Add read-only public list/detail handlers backed only by a synthetic in-memory repository; leave teacher writes and all real-source ingestion out of this slice.
- [x] Implement field filters, inclusive source-publication-date bounds, stable relevance ordering, cursor pagination, loading/empty/error/detail states and clear/back navigation.
- [x] Add automated cases for exact, alias, substring, one-edit typo and negative queries; date boundaries, invalid input, repository failure, pagination and read endpoint access/default deny.
- [x] Run applicable project checks and record the exact revision and results; broader T2/source collection and remaining 4-field date contract stay open.

## R26 institution-first student search amendment
- [x] Confirm institution directory first, retain direct search, cascade college choices by institution, and remove program/direction/keyword from student search.
- [x] Confirm estimated admissions-information publication date range; keep it separate from source publication date and event date, mark as an estimate.
- [x] Add a synthetic public institution directory list/detail repository and read-only API, including fictional application-entry URLs.
- [x] Make the institution directory the default information-search view; add school-specific detail and retain direct search as a separate view.
- [x] Add institution suggestions and school-to-college cascading selection; remove student program/direction/keyword filters.
- [x] Replace student source-publication date filter with inclusive estimated-publication date range.
- [x] Add tests for institution list/detail, cascade filtering, aliases, estimate date bounds/null handling and direct-search compatibility; run all project checks and browser checks on the same revision.

## R27 estimated-publication calendar
- [x] Confirm the calendar uses expected publication dates, not source publication or event dates.
- [x] Add a read-only month API that groups all public information entries by expected date and includes school, college, title and registration link.
- [x] Add a month calendar view with previous/next navigation, marked dates, selected-day results and an empty-day state.
- [x] Test valid/invalid months, multiple schools on one date, missing dates, read failures and browser month/day/link behavior.
- [x] Run all project checks, strict OpenSpec validation and browser checks on the same revision; update task and handoff records.
