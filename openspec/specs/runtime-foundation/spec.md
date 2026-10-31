# runtime-foundation Specification

## Purpose
Define the layered runtime, explicit configuration, health probes, safe logs and isolated startup behavior.

## Requirements

### Requirement: Isolated observable runtime
The system SHALL start from explicit configuration, expose bounded dependency health checks, deny unregistered APIs and emit safe JSON logs.

#### Scenario: S1 Healthy runtime
- **GIVEN** all configuration is supplied and MySQL authentication and Qdrant readiness succeed
- **WHEN** an anonymous client requests GET /health
- **THEN** HTTP 200 with code OK, version, UTC time, mysql and qdrant up states and a nonempty request_id is returned; one JSON request log contains timestamp, level, module, request_id, status and duration_ms

#### Scenario: S2 Missing or invalid configuration
- **GIVEN** any required variable is empty, a port is invalid, or the Qdrant URL has credentials/query/fragment or is not HTTP(S)
- **WHEN** the process starts
- **THEN** it exits nonzero before listening, naming only the invalid key

#### Scenario: S3 Dependency unavailable
- **GIVEN** either dependency is unreachable or MySQL authentication fails
- **WHEN** GET /health is requested
- **THEN** within 3 seconds it returns HTTP 503 with code DEPENDENCY_UNAVAILABLE and a down state without credentials, while GET /live returns HTTP 200

#### Scenario: S4 Unauthorized API
- **GIVEN** no business API has an explicit authorization registration
- **WHEN** any method targets /api/v1/resources
- **THEN** HTTP 401 with code UNAUTHORIZED and no resource data is returned

#### Scenario: S5 Unknown resource
- **GIVEN** the path is not registered and is outside /api/
- **WHEN** GET /unknown is requested
- **THEN** HTTP 404 with code NOT_FOUND is returned

#### Scenario: S6 Invalid request
- **GIVEN** health endpoints accept GET with no query parameters
- **WHEN** POST /health or GET /health?secret=value is requested
- **THEN** HTTP 405 METHOD_NOT_ALLOWED or HTTP 400 INVALID_ARGUMENT respectively is returned without reflecting the query

#### Scenario: S7 Boundary and logging safety
- **GIVEN** an incoming request ID contains untrusted text and dependencies exceed a 2 second deadline
- **WHEN** the server handles the request
- **THEN** a fresh server-generated request ID is returned, the probe times out, and logs never contain headers, query values, passwords or raw driver errors

#### Scenario: S8 Status page states
- **GIVEN** a visitor opens the system status page
- **WHEN** the health request is pending, healthy, empty, or fails
- **THEN** the page respectively shows checking, ready with version and dependency states, no status data with retry, or unavailable with retry; retry is disabled during loading

#### Scenario: S9 Clean reproduction
- **GIVEN** a clean local clone and its own configuration/project name/port
- **WHEN** docker compose up --build -d --wait is executed
- **THEN** the page and /health work without copying another project's data or configuration

#### Scenario: S10 Shutdown
- **GIVEN** the server is processing a health request
- **WHEN** SIGTERM is received
- **THEN** it allows in-flight requests up to 5 seconds to finish and exits without accepting new connections

#### Scenario: S11 Backend replacement
- **GIVEN** the gateway stays running and the backend container is replaced
- **WHEN** backend readiness returns and DNS cache expires within 5 seconds
- **THEN** /health resumes HTTP 200 without restarting the gateway
