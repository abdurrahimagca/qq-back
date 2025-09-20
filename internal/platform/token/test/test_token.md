# Token Module Test Plan

## Purpose & Scope
- Test JWT-based token generation and validation in `internal/platform/token`
- Verify claims structure, expiration handling, signing algorithm, and error propagation
- Ensure deterministic behaviour using environment-driven configuration

## Component Map
- **jwtTokenService (`jwt.go`)**: Implementation using `golang-jwt/jwt/v5`
- **Service Interface (`port.go`)**: Contract defining token ops
- **Domain Types (`domain.go`)**: Claims and params/results
- **Dependencies**: `environment.Environment` (Token config), `github.com/golang-jwt/jwt/v5`

## Requirements & Constraints
1. **Claims**: Include `user_id`, `sub`, `iss`, `aud`, `iat`, `exp`
2. **Expiry**: Access in minutes, Refresh in hours (from environment)
3. **Algorithm**: HS256 with shared secret; reject unexpected algs
4. **Validation**: Expired/invalid/tampered tokens return errors
5. **Security**: Use `environment.Token.Secret` for signing and verifying

## Test Strategy

### Unit Tests — jwtTokenService
- Framework: `testing`, `testify/assert`, `testify/require`
- Dependencies: Construct minimal `environment.Environment` with Token config
- No external network or clock mocking required; use tight time windows and allow small deltas

### Mock Dependencies Design
No external clients to mock. Provide helper to build `*environment.Environment` with token settings for tests.

### Test Matrix

#### Constructor
- **`NewJWTTokenService`**
  - Returns non-nil service with provided environment

#### Token Generation
- **`GenerateTokens`**
  - Returns non-empty `AccessToken` and `RefreshToken`
  - Decoded access claims match `user_id`, `sub`, `iss`, `aud`
  - Access `exp` ~ now + `AccessTokenExpireTime` minutes (±2s)
  - Refresh `exp` ~ now + `RefreshTokenExpireTime` hours (±2s)
  - Algorithm header is `HS256`

#### Token Validation
- **`ValidateToken`**
  - Valid HS256 token signed with configured secret → `Claims` returned
  - Expired token → error
  - Token signed with different secret → error
  - Token with unexpected algorithm (e.g., RS256 header) → error
  - Tampered token (payload/ signature) → error
  - Malformed token / empty string → error

### Test Utilities & Layout
```
internal/platform/token/test/
├── jwt_service_test.go           # Unit tests for Generate/Validate
└── test_helpers_test.go          # Helpers to build environment, parse claims
```

Suggested helpers:
- BuildEnv(secret, accessMins, refreshHours, issuer, audience string) *environment.Environment
- ParseClaims(token string) (*Claims, map[string]any, error)

### Edge Cases
- Empty `UserID` during generation → still produces tokens; subject should mirror `UserID`
- Audience as multiple strings (ensure claim shape)
- Very small expirations (1s) to verify expiry behaviour

### Environment Setup
- Provide ephemeral values in tests:
  - `TOKEN_SECRET`, `ACCESS_TOKEN_EXPIRE_TIME`, `REFRESH_TOKEN_EXPIRE_TIME`, `ISSUER`, `AUDIENCE`
- Construct `environment.Environment` directly in tests instead of relying on OS env

### Running The Suite
- Unit tests: `go test ./internal/platform/token/... -count=1`

### Success Criteria
- ✅ Claims contents correct and complete
- ✅ Expirations computed from environment within tolerance
- ✅ Only HS256 accepted; other algs rejected
- ✅ Invalid/expired/tampered tokens rejected

### Future Enhancements
- Add clock abstraction to avoid timing tolerances
- Add support and tests for token revocation/blacklist
- Add `aud`/`iss` claim validation toggles and tests
- Add refresh rotation semantics and tests
