# Design

## Scope
This change defines public information discovery and teacher curation: anonymous visitors and students can search/read published information; teachers can curate network-source candidates and publish approved revisions. Student profiles and teacher workspaces remain private. Positioning and simulation notes below are future design alignment only; no implementation is claimed.

## Roles and access
- **Anonymous:** read only currently published public-information list/detail. No upload, profile, personal result, consultation, teacher management or internal-knowledge access.
- **Student:** same public information reads, plus only their own private submissions/results after authentication. No teacher management or internal knowledge.
- **Teacher:** authenticated management of information and authorized access to student records/internal knowledge according to separately defined permissions.

Identity and role are resolved server-side. Public reads do not require a session. Private endpoints without an identity return 401 UNAUTHORIZED; an authenticated but disallowed role returns 403 FORBIDDEN; another student's resource returns 404 NOT_FOUND. Every handler checks authorization before reading/writing protected storage. Client-side visibility is not an access control. Login protocol, enrollment and password recovery are outside this change.

## Proposed record contract
These are explicit design recommendations for the first implementation, not claims that the owner has confirmed every value or constraint. Category taxonomy, whether a source URL must be mandatory, field limits and similarity rules remain review items. Empty scope arrays mean broad coverage; this avoids fabricating specific college/program values.

| Field | Type and rule | Required / visibility |
|---|---|---|
| `id` | UUID | System generated; public for published records |
| `institution_name` | trimmed UTF-8 string, 1–160 Unicode code points | Required; public |
| `college_names` | array of 0–20 trimmed UTF-8 strings, each 1–160 code points | Required array; empty means institution-wide information; public |
| `program_names` | array of 0–20 trimmed UTF-8 strings, each 1–120 code points | Required array; empty means no program restriction; public |
| `direction_names` | array of 0–20 trimmed UTF-8 strings, each 1–120 code points | Required array; empty means no narrower direction label; public |
| `college_aliases`, `direction_aliases` | arrays of 0–50 trimmed UTF-8 strings, each 1–120 code points | Optional; public search index only after review |
| `category` | enum, 1–32 ASCII chars; proposed values `program_overview`, `application`, `assessment`, `result`, `other` | Required; public; proposed taxonomy needs owner confirmation |
| `title` | trimmed UTF-8 string, 1–160 Unicode code points | Required; public |
| `body` | plain text, 1–20,000 Unicode code points; no raw HTML | Required; public |
| `source_title` | trimmed UTF-8 string, 1–200 code points | Optional; public when present |
| `source_url` | absolute HTTPS URL, 1–2,048 chars; reject credentials and unsupported schemes | Proposed required field; owner confirmation pending |
| `public_match_summary` | trimmed UTF-8 string, 1–240 code points | Optional; public-facing short match explanation, only published after review |
| `source_published_on` | Gregorian date `YYYY-MM-DD` | Nullable; date the source says the information was published |
| `application_deadline_on` | Gregorian date `YYYY-MM-DD` | Nullable; deadline date, not publication date |
| `event_start_on`, `event_end_on` | Gregorian dates `YYYY-MM-DD` | Independently nullable; when both are present, start must be <= end |
| `status` | `draft` / `published` / `withdrawn` | System controlled; public output only ever implies `published`, management-only otherwise |
| `revision` | positive integer, starts at 1, monotonic per record | System controlled; public detail may expose revision |
| `published_at` | UTC RFC3339 timestamp | System assigned on each successful publish; public |
| `withdrawn_at` | UTC RFC3339 timestamp or null | System assigned at withdrawal; management-only |
| `internal_note` | trimmed UTF-8 string, 0–4,000 code points | Optional; teacher-only, never in public DTO/log |
| `created_by`, `published_by`, `withdrawn_by` | authenticated teacher UUID, nullable only before the corresponding action | Management/audit-only; never public |
| `created_at`, `updated_at` | UTC RFC3339 timestamps | System assigned; `updated_at` management-only |

All string lengths count Unicode code points after trimming. Reject malformed UTF-8, empty required values, unsupported enum values, overlong fields, invalid calendar dates, reversed event intervals before persistence. No implicit timezone conversion applies to date-only fields. Timestamps are UTC and serialize with `Z`.

## Date search contract
`date_field` is required whenever `date_from` or `date_to` is supplied. Allowed fields are `source_published_on`, `application_deadline_on`, `event_start_on`, and `event_end_on`. Bounds use inclusive calendar-day comparison (`>= date_from`, `<= date_to`); a missing bound is unbounded. Missing/null values do not match a date-filtered query. If both bounds exist and `date_from > date_to`, return 400. The filter never silently switches between deadline, event and source publication dates. `published_at` is an audit timestamp and is not the default business-date filter.

## Query, order and pagination
- `institution`, `college`, `program` and `direction` are optional query terms after trimming and Unicode NFC normalization. Student public search accepts fuzzy terms against institution/college/program/direction names, reviewed aliases, title and body, with best-match sorting; a school-wide record can match a program query if its scope is broad. `keyword` is optional, 1–80 Unicode code points after trimming and searched as a literal term. Matching order (proposed) is normalized exact name, reviewed exact alias, prefix/token match, then fuzzy similarity. Similarity threshold is not fixed until representative sample review. Public results expose matched public terms and a short approved summary only. Results do not claim exhaustive institutional coverage. No filter means all currently published records, never all statuses.
- Date parameters: `date_field`, optional `date_from`, optional `date_to`; invalid combinations fail with 400.
- Sort allowlist: `relevance` (default when a search term exists, then `published_at desc`, then `id asc`), `published_at_desc` (default without a term, then `id asc`), `published_at_asc`, `title_asc` (then `id asc`). No arbitrary SQL/order expression.
- Cursor pagination only: default page size 20, allowed 1–100. The opaque cursor binds normalized filters, sort, last sort key and a 15-minute expiry; changing filters/sort with an old cursor returns 400. New publications after a page cursor are not inserted into later pages of that traversal; withdrawn records are excluded on every page. Refreshing the search starts a new traversal. This avoids offset drift while keeping withdrawals effective.
- Empty results return 200 with an empty array and no cursor. Student and anonymous list responses share the same explicit public DTO.


## Faculty search and evidence model

Student public search is faculty-oriented and fuzzy. Index fields are institution, college, program, direction, reviewed aliases, published title and public body. The proposed ranking is name exact → reviewed alias exact → prefix/token overlap → fuzzy similarity, with publication time and ID tie-breakers. The exact algorithm and threshold remain recommendations pending representative evaluation; no result implies that every relevant faculty has been indexed. Student fields include institution/college/direction, public information, matched public term, approved short explanation and public source link.

Teacher search may cross institutions and disciplines. A result is a candidate association with a typed relation (`direct` or `cross-disciplinary`), query term, matched name/alias, rationale (1–1,000 code points), source references, evidence excerpt/locator, source timestamps, conflict state, review state and reviewer. A proposed logical association record has UUID, institution, college, subject terms (1–30), aliases (0–100), relation type, rationale, 1–20 source references, status (`candidate`, `reviewed`, `published`, `withdrawn`) and verification actor/time. Evidence excerpt is limited to 1,000 code points and retains a locator into the source snapshot. Do not assign a numerical confidence unless it is empirically calibrated; show evidence and review state instead. Missing evidence means no asserted cross-disciplinary relationship. Teacher results are candidates and do not claim exhaustive coverage.

## Network source intake and curation

The planned workflow is collection task → source snapshot/metadata → extracted candidates → deduplication/category suggestions → institution/college/direction/alias associations → conflict/staleness review → teacher correction/approval → information draft revision → explicit publish/withdraw. One source may produce multiple candidate items/categories; each derived field/claim keeps a source reference and source-text locator. Exact duplicates may be flagged, not silently merged; conflicting dates or claims remain separately visible for review. Refreshing a source creates a new snapshot and candidate revision; the currently published record remains visible until a teacher publishes the reviewed replacement. A failed update leaves the current published revision unchanged.

Logical source record contract (design proposal): `source_id` UUID; `canonical_url` absolute HTTPS URL (1–2,048 chars); `title` 1–200 chars; `publisher` 1–160 chars; `fetched_at` UTC RFC3339; nullable `source_published_on` date; `content_hash` lowercase SHA-256 hex; `excerpt` 1–1,000 code points; `locator` 1–300 chars; `access_state` (`available`, `restricted`, `failed`); `snapshot_ref` points to controlled local source storage, never Git. Collection only addresses permitted public pages; restricted/authenticated content is not bypassed. The site list, schedule, adapter, raw-snapshot retention and licensing rules remain to be settled before implementation.

MCP may provide a tool interface for retrieving pages or passing source text, but its response is untrusted input and is not a quality or provenance guarantee. Store original URL, retrieval time, version/hash and evidence location independently; teacher review remains required. This change does not select a production collection connector or source schedule.

Public information DTOs and teacher curation DTOs are separate. Public DTOs exclude crawl logs, internal notes, conflict internals, student identity and unreviewed candidates. Teacher views expose richer evidence and review metadata. Public user-provided source links can appear on a published information detail; internal recommendation knowledge sources remain hidden from student positioning output.

## Follow-on product interface alignment (not in this capability slice)

- Positioning remains school-level but uses a natural, organized feedback structure around evidenced strengths, fit directions and preparation next steps; never label the student-facing card “school-level recommendation”, impersonate a real counselor, invent strengths or promise admission. A product-specific expression guide/skill can be designed separately; it does not modify global Codex skills.
- Consultation CTA text: “如果需要详细的免费的定位咨询服务，可以点击下方按钮～”; button: “领取详细免费定位咨询”. Click creates a pending consultation request with student account ID, contact source (`account_contact` or `entered_for_request`), phone snapshot, WeChat ID snapshot, timestamp and pending status. The student selects an existing account contact or enters/updates contact for this request; if no usable account contact exists, require one chosen contact value before submission. Phone and WeChat ID are distinct. No automatic friend-add or outbound message is implied.
- A single viewable privacy notice at registration or submission covers consultation, submitted-experience case curation, destination verification and contact use. Reuse that notice without repeating it in separate dialogs. Teachers can view submitted materials under server-side role control without per-submission student approval. Draft/unsubmitted browser input is not a teacher-visible submission.
- The simulation interface remains a placeholder until the reference interaction has been verified. Define its final steps only from verified reference behavior.

## Revision and publication consistency
A record has immutable numbered revisions. Editing a draft changes that draft revision. Editing an already published record creates a new draft revision while the last published revision remains visible. Publish validates and atomically switches the public pointer to the selected revision and assigns `published_at`; a failed switch leaves the prior public revision unchanged. Withdrawal atomically removes the public pointer and changes management status; old revisions remain teacher-visible but are not publicly addressable. Each write supplies the expected current revision; stale writes return 409 CONFLICT. Repeating the same idempotency key with the same operation returns the original result; reusing it for a different payload returns 409. Public detail reads one revision snapshot, never a mixture of fields across revisions.

## HTTP and error contract
Public endpoints: `GET /api/v1/public-information` and `GET /api/v1/public-information/{id}`; both allow anonymous and student identities. Management endpoints are authenticated teacher-only. Student private endpoints (for example profile upload and personal results) require an authenticated student. The exact private route list belongs to their later capability contracts.

Responses use `{code,message,data,request_id}`. Public success is HTTP 200 / `OK`; create is 201 / `CREATED`; invalid input, filter, cursor or revision is 400 / `INVALID_ARGUMENT`; no identity on a private endpoint is 401 / `UNAUTHORIZED`; wrong role is 403 / `FORBIDDEN`; missing, withdrawn or foreign private resource is 404 / `NOT_FOUND`; stale revision or idempotency conflict is 409 / `CONFLICT`; storage unavailable is 503 / `DEPENDENCY_UNAVAILABLE`. New codes must be registered before use. Responses/logs contain no credentials, raw storage errors, internal notes, private profile data or full request bodies.

## UI and status behavior
Visitors and students share the same four student navigation sections and may browse each page without signing in. Identity-related actions such as uploading a resume, saving/submitting personal background, reading personal records/results or creating an account-linked consultation prompt login/registration and return to the originating section with the current operation context. Private server endpoints still require authentication and ownership checks; page visibility never grants access. Public list/detail retain loading, results, empty and failure states. Teacher operations and internal knowledge require teacher authorization. The simulation remains a video-pending placeholder. All screens retain keyboard focus, visible labels, narrow-screen reflow and obvious primary actions per product.md.

## Verification mapping
- OpenSpec delta scenarios define public anonymous success, unpublished invisibility, private anonymous denial, role/ownership rejection, date/filter and lifecycle behavior.
- Implementation later maps these to handler/service/storage tests, UI state tests and browser checks, using synthetic data only.
- Existing runtime-foundation, extension-contract, release-verification and eight engineering standards remain in force. No existing base specification is replaced; no standards exception is proposed.




### Navigation and account entry design follow-up
The public information search remains the anonymous default entry. A student account view is organized into four separate top-level sections: information search, background positioning, system-filling simulation, and personal center; personal profile is nested under personal center. Teacher workspace has a separate entry and navigation. Registration/login may prioritize exploring WeChat as a provider, but provider eligibility, binding flow, and authentication service are unverified and are not selected or integrated by this change. The simulation login gate remains pending reference-flow verification. These are interface design notes only; they do not extend this capability slice into account implementation.

## First S1 implementation slice
The product owner selected the existing web client with synthetic records. The UI lives at `/information`; `/` and `/status/details` remain the required runtime-foundation status pages. Information search contains two views: an institution directory, shown first, and retained direct search. The directory lists institutions and opens institution-level details with college information and an application entry. Direct search offers institution-name suggestions; selecting an institution constrains the college choices to that institution. Student direct search filters only institution and college; program, direction and keyword filters are not shown to students and are reserved for a future teacher capability. The public read API uses `GET /api/v1/public-information` and `GET /api/v1/public-information/{id}` plus institution directory list/detail reads, all backed only by a read-only in-memory synthetic repository. All institution names, application entries and information are fictional demo content; no real admission links or real source claims are introduced. This slice adds no database migration, teacher curation, real source collection, student data, model call, or authentication provider.

The student date-range filter uses the estimated publication date for the admissions information (for example, when a summer-camp or pre-recommendation notice may be issued). It does not filter the real source publication date or the camp/event date. Each record may have a nullable `expected_publication_on`; inclusive filter bounds match that estimate and records without an estimate do not match a dated query. The UI labels this value as estimated. Synthetic dates are demonstration values, not verified school forecasts. The broader capability contract for source publication, application deadlines, event dates and platform publication timestamps remains distinct and unchanged.

The institution directory uses `GET /api/v1/public-institutions` and `GET /api/v1/public-institutions/{id}`. The list supports browsing; an institution detail includes its colleges and public admission entries. Student direct search remains available with institution-name suggestions; a selected institution constrains the college options. Student filters exclude program, direction and keyword; those criteria belong to a later teacher search capability. This slice uses fictional institutions and reserved `.invalid` demonstration links. Replacing them with real admission links requires a separate approved data intake/teacher-publication slice.

The information-search views also include a month calendar backed by `GET /api/v1/public-information/calendar?month=YYYY-MM`. The calendar groups public admission entries by nullable `expected_publication_on`, and each day shows its institution, college, title and registration link. Selecting an empty day shows an empty state. Month navigation is limited to valid Gregorian months; dates and links remain synthetic in this slice. The calendar is not a schedule of actual source publication or event dates.

Search first ranks exact normalized names, exact aliases, prefixes and substrings. It then allows a one-code-point Damerau-Levenshtein edit (substitution, insertion/deletion, or adjacent transposition) against a candidate window whose length may differ by one code point, for query terms of at least four Unicode code points. Shorter queries do not receive typo tolerance. Search fields are institution, college, program, direction, aliases, title and body. Match tier determines relevance; ties use platform publication time descending and stable ID order. Normalization trims and lowercases Unicode text; no extra normalization library is introduced in this slice. A synthetic positive/negative query set calibrates obvious false matches; it does not establish real-world recall or exhaustive coverage. All records, including aliases, are fictional fixtures; source previews are clearly marked synthetic and do not claim real provenance.
