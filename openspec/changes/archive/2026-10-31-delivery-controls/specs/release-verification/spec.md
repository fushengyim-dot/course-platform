## ADDED Requirements
### Requirement: Isolated release verification
The project SHALL verify a selected revision in an isolated Compose project and report health, revision and elapsed time without exposing credentials.

#### Scenario: Fresh start
- **GIVEN** a valid Git revision and an unused local port
- **WHEN** the verification script runs
- **THEN** it exports that revision, builds and starts all services, verifies health and stops only its own project

#### Scenario: Rollback
- **GIVEN** two distinct version references
- **WHEN** rollback verification runs
- **THEN** previous, current and previous revisions each become healthy in sequence and their durations are recorded

#### Scenario: Failure cleanup
- **GIVEN** a build or health failure
- **WHEN** verification ends
- **THEN** it exits nonzero, records the failure without passwords and attempts to stop its own project

### Requirement: Protected main branch
The repository SHALL require a Pull request and successful automated verification for main updates.

#### Scenario: Direct push
- **WHEN** a new commit is pushed directly to main
- **THEN** the remote rejects it without changing main

#### Scenario: Verified pull request
- **GIVEN** a Pull request with passing checks
- **WHEN** it is merged
- **THEN** main contains the reviewed commits and the recorded check result
