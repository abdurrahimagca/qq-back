# Registration Module Test Plan

## Purpose & Scope
- Cover OTP-based registration/login and token refresh flows in `internal/registration`
- Focus on bad requests (empty/invalid inputs), invalid refresh tokens, and `isNewUser` behaviour
- Mail delivery is covered in mailer tests; here we assert only that mailer is invoked correctly

## Component Map
- **Use case (`registration.service.go`)**: `registrationUsecase`
  - `RegisterOrLoginOTP(ctx, email) (*bool, error)`
  - `VerifyOTPAndLogin(ctx, email, otp) (GenerateTokenResult, error)`
  - `RefreshTokens(ctx, refreshToken) (GenerateTokenResult, error)`
- **Server (`registration.server.go`)**: `registrationServer`
  - Handlers mapping to Huma operations: Send OTP, Verify OTP, Refresh Tokens
- **Dependencies**
  - `auth.Service` (OTP gen/verify, kill orphaned otps, WithTx)
  - `user.Service` (lookup/create, WithTx)
  - `token.Service` (generate/validate tokens)
  - `mailer.Service` (GetTemplate/SendEmail) — behaviour tested elsewhere
  - `*pgxpool.Pool` (transactions)

## Requirements & Behaviours
1. **Send OTP**
   - Existing user → `isNewUser=false`; new user created on first-time login → `isNewUser=true`
   - Kills orphan OTPs, generates and saves a new OTP
   - Retrieves `otp` template and sends email with OTP inserted
2. **Verify OTP**
   - Verifies provided OTP; cleans up orphan OTPs; returns token pair
3. **Refresh Tokens**
   - Validates refresh token; user id must be present and valid UUID; returns new token pair
4. **Errors**
   - Propagate underlying service/DB errors
   - Map to Huma errors in server layer via `qqerrors.GetHumaErrorFromError`

## Test Strategy
- Unit tests for use case with mocks for all dependencies (no DB or network)
- Unit tests for server handlers using a fake `RegistrationUsecase`
- Input validation of struct tags is primarily an HTTP concern; we still cover empty strings flowing to usecase and verify errors propagate

## Mocks & Fakes
- **Auth mock**: `WithTx(pgx.Tx)`, `CreateNewAuthForOTPLogin`, `GenerateAndSaveOTPForAuth`, `VerifyOTP`, `KillOrphanedOTPsByUserID`
- **User mock**: `WithTx(pgx.Tx)`, `GetUserByEmail`, `CreateDefaultUserWithAuthID`, `GetUserByID`
- **Token mock**: `GenerateTokens`, `ValidateToken`
- **Mailer mock**: `GetTemplate`, `SendEmail` (only interaction verified)
- **DB pool/Tx mock**: Provide a fake `pgxpool.Pool` with `Begin(ctx)` returning a fake `pgx.Tx` capturing `Commit`/`Rollback` calls (or use a small wrapper interface in tests)

## Test Matrix (Use Case)

### RegisterOrLoginOTP(ctx, email)
- Happy path — existing user
  - `GetUserByEmail` → existing; `KillOrphanedOTPsByUserID`; `GenerateAndSaveOTPForAuth` → code
  - `mailer.GetTemplate("otp")`, `mailer.SendEmail(...)`
  - Commits transaction; returns `isNewUser=false`
- Happy path — new user
  - `GetUserByEmail` → not found; `CreateNewAuthForOTPLogin`; `CreateDefaultUserWithAuthID`
  - `KillOrphanedOTPsByUserID`; generate OTP; mailer calls; commit; `isNewUser=true`
- Errors
  - Begin fails → error
  - `GetUserByEmail` returns unexpected error → error
  - `CreateNewAuthForOTPLogin` fails → error
  - `CreateDefaultUserWithAuthID` fails → error
  - `KillOrphanedOTPsByUserID` fails → error
  - `GenerateAndSaveOTPForAuth` fails → error
  - `mailer.GetTemplate` fails → error (mailer behaviour itself tested elsewhere)
  - Commit fails → error
  - `mailer.SendEmail` fails → error
- Bad requests
  - Empty email → treat as normal input; expect downstream (user/auth) to return validation error; ensure it propagates

### VerifyOTPAndLogin(ctx, email, otp)
- Happy path
  - `VerifyOTP` success; fetch user; kill orphan OTPs; commit; `GenerateTokens` → tokens
- Errors
  - Begin fails → error
  - `VerifyOTP` fails (invalid/expired/empty code) → error
  - `GetUserByEmail` fails → error
  - `KillOrphanedOTPsByUserID` fails → error
  - Commit fails → error
  - `GenerateTokens` fails → error
- Bad requests
  - Empty otp string → `VerifyOTP` returns validation error; ensure it propagates

### RefreshTokens(ctx, refreshToken)
- Happy path
  - `ValidateToken` returns claims with `UserID`; parse UUID; `GetUserByID`; `GenerateTokens` → new tokens
- Errors
  - `ValidateToken` returns error (invalid refresh token) → error
  - Claims `UserID` empty → `qqerrors.ErrValidationError`
  - Invalid UUID in claims → error
  - `GetUserByID` fails → error
  - `GenerateTokens` fails → error
- Bad requests
  - Empty refresh token → `ValidateToken` returns error; ensure it propagates

## Test Matrix (Server Handlers)
- `SendOtpHandler`
  - Success returns body with `isNewUser`
  - Usecase error mapped via `qqerrors.GetHumaErrorFromError`
- `VerifyOtpHandler`
  - Success returns tokens
  - Empty `otpCode`/invalid → usecase returns error; verify mapping
- `RefreshTokensHandler`
  - Success returns tokens
  - Invalid/empty refresh token → error mapping verified

## Test Utilities & Layout
```
internal/registration/test/
├── usecase_test.go         # Unit tests for registrationUsecase with mocks
├── server_test.go          # Handler tests with fake usecase
├── mocks.go                # Mocks for auth, user, token, mailer, db/tx
└── test_helpers.go         # Builders for users/claims, tx harness
```

## Edge Cases
- Email with leading/trailing spaces (trim at caller? ensure downstream error or normalization)
- Multiple OTP generations quickly: ensure old OTPs are killed each time
- Idempotency: requesting OTP repeatedly should not error

## Running
- Use case tests: `go test ./internal/registration/test -run Usecase -count=1`
- Handler tests: `go test ./internal/registration/test -run Server -count=1`
- Full: `go test ./internal/registration/test -count=1`

## Success Criteria
- ✅ `isNewUser` correctness for existing/new users
- ✅ Empty/invalid codes and refresh tokens produce errors that propagate
- ✅ All critical dependencies are invoked with expected parameters
- ✅ Transactions commit/rollback appropriately across success/error paths