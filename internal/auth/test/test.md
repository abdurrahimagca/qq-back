# Auth Module Test Plan

## Purpose & Scope
- Drive consistent, high-value tests for `internal/auth`
- Cover business rules in `auth.service.go` and persistence logic in `auth.repo.go`
- Exclude `auth.domain.go` (constant error declarations only)

## Component Map
- **Service (`auth.service.go`)**: generates OTP codes, delegates persistence to `Repository`, enforces email ↔ OTP matching, exposes cleanup helpers, and allows transactional composition through `WithTx`.
- **Repository (`auth.repo.go`)**: wraps SQLC queries; converts database failures into `qqerrors`. Requires end-to-end verification with a real PostgreSQL instance because it is a thin layer over generated code.

## Unit Tests — Service Layer
- Framework: `testing`, `testify/assert`, `testify/require`, optional `testify/suite` for shared fixtures.
- Dependencies: in-memory fake implementing `auth.Repository`; injectable `io.Reader` for deterministic OTP generation (override `rand.Reader` inside tests).

### Fake Repository Design
- Struct keeps OTPs in maps keyed by auth ID and OTP hash; tracks emails per auth ID.
- `WithTx` returns the same fake (no real transactional semantics but must record the fact it was called for verification).
- Provide helpers: `CreateUser(authID, email)`, `GetStoredOTP(authID)` for test assertions.

### Test Matrix
- **`GenerateAndSaveOTPForAuth`**
  - Deterministic reader returns known byte slice; assert
    - Returned code equals uppercased hex of bytes
    - Stored hash equals SHA-256 of returned code
  - Error path when `rand.Read` fails (swap reader with failing stub)
  - Propagate repository errors (fake returns injected error).
- **`VerifyOTP`**
  - Happy path: fake repo returns matching email → no error
  - Email mismatch triggers `ErrInvalidOtpCode`
  - Repository returns `ErrNotFound` → expect passthrough of error
  - Repository generic failure propagates unchanged.
- **`CreateNewAuthForOTPLogin`**
  - Happy path returns ID from repository; ensure passthrough
  - Repository failure bubbles up.
- **`KillOrphanedOTPs` / `KillOrphanedOTPsByUserID`**
  - Verify delegation (fake toggles flags); error propagation.
- **`WithTx`**
  - Fake repository records `WithTx` invocation and the argument `pgx.Tx`; ensure returned service uses new repo instance; subsequent calls go through transactional fake.

### Edge-Case Considerations
- Empty email strings should surface repository validation errors (if any); ensure tests cover the behavior the repository currently enforces.
- OTP codes are case-insensitive externally—service forces uppercase; assert this in tests.

## Integration Tests — Repository Layer
- Framework: `testing`, `testify/require`; manage container lifecycle with `testcontainers-go`.
- Setup steps per suite:
  1. Start PostgreSQL container (matching project version).
  2. Run migrations from `db/migrations` (use Migrator already wired for tests or apply via `db` package helpers).
  3. Create pooled connection (`pgxpool.New`).
  4. Instantiate repository with pool and clean tables before each test.
- Teardown: close pool, terminate container.

### Repository Test Matrix
- **`CreateAuthForOTPLogin`**
  - Inserts row with expected provider `email_otp`; verify record via SQLC query.
  - Duplicate email constraint returns converted `qqerrors.ErrConflict` (depending on schema) — assert error type.
- **`CreateOTP`**
  - Persists hashed code; verify presence and foreign-key relation to auth row.
- **`GetUserIdAndEmailByOtpCode`**
  - Happy path returns matching row.
  - Missing OTP → expect `auth.ErrNotFound`.
  - Database error simulation (e.g., close pool to force failure) → expect wrapped `qqerrors`.
- **`KillOrphanedOTPs` / `KillOrphanedOTPsByUserID`**
  - Seed multiple OTPs; assert targeted deletions.
  - Concurrency: run deletion in parallel with insertion to ensure no panics (use subtests with `t.Parallel`).
- **`WithTx`**
  - Acquire explicit transaction; call repository methods through transactional repo; assert data committed/rolled back when transaction is committed/rolled back manually in test.

## Test Utilities & Layout
```
internal/auth/test/
├── service_test.go        // Service unit tests using fake repository
├── fake_repository.go     // In-memory fake + helpers described above
├── integration_test.go    // Repository integration tests with testcontainers
└── test_helpers.go        // Shared setup (migrations, deterministic rand reader)
```
- Keep fake and helpers non-exported; use Go build tag `//go:build test` if helpers should not leak into production builds.

## Running The Suite
- Unit only: `go test ./internal/auth/test -run Service`
- Integration: `go test -tags=integration ./internal/auth/test -run Integration`
- Full project: `make test`

## Future Enhancements
- Add metrics/assertions when rate limiting or throttling is introduced.
- Once email throttling logic exists, extend fake repo to track timestamps for OTP regeneration limits.
