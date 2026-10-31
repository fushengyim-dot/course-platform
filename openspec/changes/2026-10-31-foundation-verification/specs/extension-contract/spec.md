## ADDED Requirements
### Requirement: Reusable status pages
The system SHALL render the home and status-details pages with a shared status component and consistent state messages.

#### Scenario: Shared page rendering
- **GIVEN** a healthy dependency response
- **WHEN** the user opens / or /status/details
- **THEN** both pages show the shared status component with version, time and dependency states; each page identifies its purpose and provides navigation

#### Scenario: Failed or empty health response
- **GIVEN** a failed, malformed or empty health response
- **WHEN** either page loads
- **THEN** the common error or empty message and retry control are displayed without exposing raw errors

#### Scenario: Unknown page
- **WHEN** an unregistered page is requested
- **THEN** HTTP 404 is returned

### Requirement: Isolated read and write contract
The example SHALL expose authenticated GET and POST record handlers with layered dependencies and bounded input, using isolated test storage.

#### Scenario: Create and read
- **GIVEN** an authenticated subject and a valid title with an idempotency key
- **WHEN** a record is created and then read by its id
- **THEN** the write returns 201 and the read returns 200 using code, message, data and request_id fields

#### Scenario: Invalid input
- **WHEN** title is blank or longer than 80 characters, the key is missing, JSON is malformed, unknown fields or trailing JSON are supplied, or the body exceeds 1024 bytes
- **THEN** HTTP 400 is returned and no record is written

#### Scenario: Unauthorized subject
- **WHEN** a request has no valid identity
- **THEN** HTTP 401 is returned without reading or writing storage

#### Scenario: Missing or foreign record
- **WHEN** a record is missing or belongs to another subject
- **THEN** HTTP 404 is returned

#### Scenario: Dependency failure
- **GIVEN** the repository cannot read or write
- **WHEN** an authenticated valid request is handled
- **THEN** HTTP 503 is returned without storage details

#### Scenario: Idempotency and concurrency
- **GIVEN** concurrent requests with the same subject, key and title
- **WHEN** they create a record
- **THEN** they return the same record id and storage contains one record; a different title with that subject and key returns 409

#### Scenario: Production isolation
- **WHEN** the production server receives /api/v1/records requests
- **THEN** its existing HTTP 401 behavior remains unchanged
