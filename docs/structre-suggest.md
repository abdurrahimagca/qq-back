your-app/
├── cmd/
│   ├── api/
│   │   └── main.go           # API server entry point
│   ├── worker/
│   │   └── main.go           # Background worker
│   └── migrate/
│       └── main.go           # Migration runner
├── internal/
│   ├── database/
│   │   ├── connection.go     # DB connection setup
│   │   ├── queries/          # Generated sqlc code
│   │   └── migrations/       # SQL migration files
│   ├── domain/               # Business entities/models
│   │   ├── user.go
│   │   ├── review.go
│   │   ├── media.go
│   │   └── message.go
│   ├── service/              # Business logic
│   │   ├── auth/
│   │   ├── user/
│   │   ├── review/
│   │   ├── messaging/
│   │   └── media/
│   ├── handler/              # HTTP/WebSocket handlers
│   │   ├── rest/
│   │   │   ├── auth.go
│   │   │   ├── user.go
│   │   │   └── review.go
│   │   ├── websocket/
│   │   │   └── messaging.go
│   │   └── middleware/
│   │       ├── auth.go
│   │       ├── privacy.go
│   │       └── ratelimit.go
│   ├── config/
│   │   └── config.go         # Configuration management
│   ├── crypto/
│   │   ├── encryption.go
│   │   └── otp.go
│   └── external/             # Third-party integrations
│       ├── email/
│       ├── oauth/
│       ├── tmdb/
│       └── storage/
├── pkg/                      # Reusable utilities
│   ├── logger/
│   ├── validator/
│   ├── jwt/
│   └── errors/
├── sql/
│   ├── migrations/           # Migration files
│   │   ├── 001_initial.up.sql
│   │   ├── 001_initial.down.sql
│   │   ├── 002_privacy.up.sql
│   │   └── 002_privacy.down.sql
│   ├── queries/              # SQL queries for sqlc
│   │   ├── users.sql
│   │   ├── reviews.sql
│   │   ├── media.sql
│   │   └── messages.sql
│   └── schema.sql            # Complete schema (for reference)
├── docker/
│   ├── api.Dockerfile
│   ├── worker.Dockerfile
│   └── nginx.conf
├── docker-compose.yml
├── docker-compose.dev.yml    # Development overrides
├── sqlc.yaml                 # sqlc configuration
├── Makefile                  # Common commands
├── .env.example
└── README.md