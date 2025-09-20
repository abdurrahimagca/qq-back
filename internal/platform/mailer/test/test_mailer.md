# Mailer Module Test Plan

## Purpose & Scope
- Test email sending and template retrieval in `internal/platform/mailer`
- Verify integration with Resend client and embedded templates
- Ensure input validation and error propagation

## Component Map
- **resendMailer (`resend.go`)**: Implementation using `resend-go`
- **Service Interface (`port.go`)**: Contract defining mailer ops
- **Templates (`templates/*.html`)**: Embedded HTML templates via `embed.FS`
- **Dependencies**: `environment.Environment`, `github.com/resend/resend-go/v2`, Go `embed`

## Requirements & Constraints
1. **Sending**: Sends HTML emails with given from/to/subject/body
2. **Templates**: Retrieve by name; prevent path traversal
3. **Errors**: Network/Resend errors are surfaced to caller
4. **Security**: Disallow `..` sequences in template names

## Test Strategy

### Unit Tests — resendMailer
- Framework: `testing`, `testify/assert`, `testify/require`
- Use a mock Resend client to avoid network calls
- Test embedded template retrieval from `embed.FS`

### Mock Dependencies Design

#### Mock Resend Client
- Minimal mock covering `Emails.Send(*resend.SendEmailRequest) (*resend.SendEmailResponse, error)`
- Track inputs and allow configuring outcomes

#### Test Environment
- Construct minimal `environment.Environment` with `Resend.Key`

## Test Matrix

### Constructor
- **`NewResendMailer`**
  - Returns service with non-nil client and env

### Template Retrieval
- **`GetTemplate`**
  - Existing template (e.g., `otp`) returns non-empty HTML
  - Non-existing template returns error
  - Template name with `..` returns error

### Sending
- **`SendEmail`**
  - Happy path calls client with expected fields
  - Propagates client error
  - Handles empty/whitespace body (still sends)

## Test Utilities & Layout
```
internal/platform/mailer/test/
├── resend_service_test.go   # Unit tests with mocks
└── mocks.go                 # Simple mock client for resend
```

Suggested Mock Shape:
```go
// EmailsAPI is the subset we use; define an interface for test injection
 type EmailsAPI interface {
     Send(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error)
 }
```

## Edge Cases
- `From` or `To` not set (decide whether to validate or rely on Resend)
- Large HTML body
- Unicode and special characters in subject/body

## Environment Setup
- Construct `environment.Environment` inline in tests; no OS env dependency

## Running The Suite
- Unit tests: `go test ./internal/platform/mailer/... -count=1`

## Success Criteria
- ✅ Templates resolved correctly and safely
- ✅ Email send requests formed correctly
- ✅ Client/network errors propagated

## Future Enhancements
- Add text body support and tests
- Add templating with variables and tests
- Add rate limiting/retry logic tests
