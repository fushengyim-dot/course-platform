# Published Information Specification

## Purpose
Define public, login-free information discovery and evidence-backed teacher curation. Student profiles and teacher workspaces remain protected.

## ADDED Requirements

### Requirement: Teacher-managed information lifecycle
The system SHALL allow an authenticated teacher to create/edit drafts, explicitly publish information, revise published information with traceable history, and withdraw it. Anonymous users and students SHALL NOT perform management actions.

#### Scenario: Save a draft
- **GIVEN** an authenticated teacher submits valid information fields
- **WHEN** the teacher saves without publishing
- **THEN** the system stores a draft and excludes it from every public list and detail response

#### Scenario: Invalid information fields
- **GIVEN** a teacher submits a missing required field, invalid category, overlong string, unsafe source URL, invalid date, or reversed event interval
- **WHEN** the teacher saves or publishes the record
- **THEN** the system returns HTTP 400 INVALID_ARGUMENT and does not change the stored draft or public revision

#### Scenario: Publish information
- **GIVEN** an authenticated teacher has a valid draft with its source and evidence fields
- **WHEN** the teacher explicitly publishes it
- **THEN** the system atomically records the published revision and makes that revision available to public queries

#### Scenario: Failed publication
- **GIVEN** publication storage fails
- **WHEN** a teacher publishes a valid draft
- **THEN** the system reports a safe failure and leaves the previous public revision unchanged, or keeps a new draft unavailable

#### Scenario: Update keeps the previous public revision until publish
- **GIVEN** a record has a currently published revision
- **WHEN** a teacher edits content from a fresh source snapshot and saves a draft revision
- **THEN** public reads continue to return the previous published revision until the teacher publishes the new one

#### Scenario: Stale revision and idempotent replay
- **GIVEN** two teacher edits use the same expected revision, or an operation is repeated with an idempotency key
- **WHEN** the stale edit or repeated operation is processed
- **THEN** the stale edit returns HTTP 409 CONFLICT; an identical replay returns the original outcome, and a different payload with the same key returns HTTP 409 CONFLICT

#### Scenario: Withdraw information
- **GIVEN** a published record
- **WHEN** an authorized teacher withdraws it
- **THEN** it disappears from public list/detail while its revision and management history remain available to teachers

### Requirement: Traceable network source intake
The system SHALL preserve provenance for information gathered from permitted public sources, derive candidate items without losing source links, and require teacher review before publication. A network response or MCP tool response SHALL NOT by itself count as verification.

#### Scenario: Capture source provenance
- **GIVEN** a collection job reads an accessible public source
- **WHEN** it records source material
- **THEN** it stores canonical URL, title, publisher when known, retrieval time, source publication date when present, content hash/version, a controlled snapshot reference and a locator to supporting text

#### Scenario: Restricted source
- **GIVEN** a source requires authentication, blocks access, or is otherwise restricted
- **WHEN** a collection job cannot access it through permitted means
- **THEN** the job records restricted/failed status and does not bypass access controls or invent extracted content

#### Scenario: One source yields multiple candidates
- **GIVEN** a source contains several distinct information items or categories
- **WHEN** extraction produces multiple candidates
- **THEN** each candidate retains its own source reference and excerpt locator, and the original source remains recoverable

#### Scenario: Duplicate or conflicting sources
- **GIVEN** two source snapshots repeat or conflict on a date, faculty/direction association, or claim
- **WHEN** the candidates are prepared for review
- **THEN** duplicate and conflict indicators are shown to the teacher; differences are not silently merged or resolved as fact

#### Scenario: Teacher review gates publication
- **GIVEN** an extracted candidate has not been reviewed
- **WHEN** a public visitor searches or opens its record
- **THEN** the candidate is not publicly visible; only a teacher may correct, reject, or turn it into a draft and explicitly publish it

#### Scenario: Source update fails
- **GIVEN** a published record has a newer failed or conflicting source snapshot
- **WHEN** refresh processing completes
- **THEN** the existing public revision remains unchanged and the failure/conflict is visible in teacher review

### Requirement: Public faculty-level fuzzy search
The system SHALL allow anonymous visitors and students to search and read published information without login. Results SHALL be organized primarily by institution and faculty/college. Student direct search SHALL filter by institution and college, suggest institutions as the visitor types, and constrain college choices to the selected institution. Public matching SHALL consider institution, college, their reviewed aliases, title and body fields. Program and direction filters are reserved for a future teacher search capability. Results SHALL show the public matched term or concise approved reason and SHALL NOT promise exhaustive coverage.

#### Scenario: Browse the public institution directory
- **GIVEN** a visitor has no authenticated session
- **WHEN** the visitor opens the institution directory and selects an institution
- **THEN** the system shows the institution's public details, colleges and associated public admission entries without requiring login

#### Scenario: Direct student search offers institution and college choices
- **GIVEN** a student or visitor opens direct public search
- **WHEN** they type an institution name or select an institution
- **THEN** the interface suggests matching institutions and limits the college choices to the selected institution; student filters include institution and college, while program, direction and keyword filters remain outside student search

#### Scenario: Anonymous visitor lists published information
- **GIVEN** a visitor has no authenticated session
- **WHEN** the visitor opens the public information home or list
- **THEN** the system returns HTTP 200 with published records only and shows useful information immediately without a login gate

#### Scenario: Anonymous visitor opens published information
- **GIVEN** a visitor has no authenticated session and a record is published
- **WHEN** the visitor requests its detail
- **THEN** the system returns HTTP 200 with only public fields for the current published revision and its public source link

#### Scenario: Fuzzy search by institution, college or alias
- **GIVEN** a visitor or student searches using a partial institution/college name or reviewed alias
- **WHEN** the public search runs
- **THEN** it returns matching published faculty-level information with the matched public term or approved concise reason, ordered by the approved matching strategy

#### Scenario: Combine and clear filters
- **GIVEN** published records match selected institution, college, program and typed date criteria
- **WHEN** a visitor combines filters then clears one or all filters
- **THEN** the result list reflects the remaining criteria and never includes drafts or withdrawn records

#### Scenario: Search does not claim completeness
- **GIVEN** the search returns one or more relevant faculty candidates
- **WHEN** results are displayed
- **THEN** the interface does not claim that all relevant institutions or faculties have been found

#### Scenario: Empty result
- **GIVEN** no published record matches valid filters
- **WHEN** an anonymous visitor or student searches
- **THEN** the system returns HTTP 200 with an understandable empty state and a way to change/clear filters

#### Scenario: Invalid search or cursor
- **GIVEN** a query, date range, sort option, page size or cursor is malformed or outside its allowed range
- **WHEN** an anonymous visitor or student searches
- **THEN** the system returns HTTP 400 INVALID_ARGUMENT and does not silently broaden the query

#### Scenario: Date filter has explicit meaning
- **GIVEN** a date range is supplied
- **WHEN** a visitor searches by the estimated publication date of admissions information
- **THEN** the system filters the nullable estimate with inclusive bounds, labels it as an estimate, and does not substitute the source publication date or event date

#### Scenario: Browse estimated publication dates on a calendar
- **GIVEN** a visitor opens the information calendar for a valid month
- **WHEN** they select a day with estimated admissions information
- **THEN** the calendar lists every matching school, college, information title and public registration link for that estimated date, clearly labelled as an estimate

#### Scenario: Calendar empty day and invalid month
- **GIVEN** a visitor selects a day with no estimated publication entries, or requests a malformed month
- **WHEN** the calendar is displayed or queried
- **THEN** the page shows a clear empty state for the day, and the API rejects a malformed month with HTTP 400 INVALID_ARGUMENT

#### Scenario: Missing, draft or withdrawn detail
- **GIVEN** a record is missing, draft or withdrawn
- **WHEN** an anonymous visitor or student requests its detail
- **THEN** the system returns HTTP 404 NOT_FOUND without revealing whether a private or withdrawn record exists

#### Scenario: Public repository failure
- **GIVEN** the public information repository is unavailable
- **WHEN** an anonymous visitor requests a list or detail
- **THEN** the system returns HTTP 503 DEPENDENCY_UNAVAILABLE without storage details

#### Scenario: Source link is unavailable
- **GIVEN** a visitor opens a published detail whose public source cannot be reached
- **WHEN** the source link fails to open
- **THEN** the interface explains that the source is unavailable and preserves navigation back to results

### Requirement: Teacher cross-institution search with evidence
The system SHALL allow an authorized teacher to search across institutions by program or direction and return direct or cross-disciplinary faculty candidates with traceable evidence and matching reasons. A cross-disciplinary relationship without supporting evidence SHALL NOT be asserted as fact.

#### Scenario: Cross-institution direct match
- **GIVEN** a teacher searches a program/direction present in reviewed evidence from more than one institution
- **WHEN** the teacher runs the search
- **THEN** each candidate shows institution, faculty, matched name/alias, relation type, rationale, source/time, excerpt locator and review/conflict state

#### Scenario: Cross-disciplinary candidate
- **GIVEN** evidence supports an association between a search direction and a faculty in another discipline
- **WHEN** an authorized teacher searches that direction
- **THEN** the faculty may appear as a cross-disciplinary candidate with its evidence and rationale clearly shown

#### Scenario: No evidence for association
- **GIVEN** a possible faculty association has no traceable supporting source
- **WHEN** teacher search results are prepared
- **THEN** the system does not present the association as a verified match

#### Scenario: Teacher search is not exhaustive
- **GIVEN** teacher search returns candidate faculties
- **WHEN** the results are displayed
- **THEN** the interface presents them as evidence-backed candidates and does not claim complete coverage of all institutions

### Requirement: Private resources and management authorization
The system SHALL enforce identity, role, publication state and resource access on the server for every operation. Public information reads SHALL NOT grant access to private student resources, teacher management or internal knowledge.

#### Scenario: Anonymous access to private student resources
- **GIVEN** a request has no authenticated identity
- **WHEN** it requests a profile upload, private submission, personal result, or consultation resource
- **THEN** the system returns HTTP 401 UNAUTHORIZED and performs no protected read or write

#### Scenario: Anonymous access to teacher resources
- **GIVEN** a request has no authenticated identity
- **WHEN** it requests teacher management or internal knowledge
- **THEN** the system returns HTTP 401 UNAUTHORIZED and reads no protected data

#### Scenario: Student cannot access teacher management or internal knowledge
- **GIVEN** an authenticated student
- **WHEN** the student requests a teacher management operation or internal knowledge resource
- **THEN** the system returns HTTP 403 FORBIDDEN and returns no protected data

#### Scenario: Student cannot access another student's private resource
- **GIVEN** an authenticated student requests a resource owned by another student
- **WHEN** the server checks resource ownership
- **THEN** it returns HTTP 404 NOT_FOUND without revealing the resource's existence

#### Scenario: Student-facing representation isolation
- **GIVEN** a teacher record contains internal notes, crawl metadata or private identity data
- **WHEN** an anonymous visitor or student reads its published representation
- **THEN** only the allowlisted public fields are returned

#### Scenario: Client visibility does not grant access
- **GIVEN** a user directly calls a protected endpoint regardless of which controls are visible
- **WHEN** the server evaluates the request
- **THEN** identity, role and ownership checks run before protected storage access
