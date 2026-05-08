## ADDED Requirements

### Requirement: Users service owns auth sessions

The users service MUST own login, refresh, and logout session behavior.

#### Scenario: Login is performed

- **WHEN** a client authenticates with valid credentials
- **THEN** the service SHALL issue access and refresh tokens

#### Scenario: Refresh is performed

- **WHEN** a client presents a valid refresh token
- **THEN** the service SHALL rotate the refresh session and issue new tokens

#### Scenario: Logout is performed

- **WHEN** a client logs out with a valid refresh token
- **THEN** the service SHALL revoke the matching refresh session

### Requirement: Refresh sessions are stored in PostgreSQL

The users service MUST persist refresh-token sessions in PostgreSQL for revocation and rotation.

#### Scenario: Session is created

- **WHEN** a user logs in
- **THEN** the service SHALL store a refresh session row in the database

#### Scenario: Session is revoked

- **WHEN** a user logs out or refreshes a session
- **THEN** the service SHALL update the stored session state accordingly
