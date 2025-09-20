package registration_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/abdurrahimagca/qq-back/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type deterministicReader struct {
	data []byte
	err  error
}

func (d *deterministicReader) Read(p []byte) (int, error) {
	if d.err != nil {
		return 0, d.err
	}
	if len(d.data) < len(p) {
		copy(p, d.data)
		return len(d.data), nil
	}
	copy(p, d.data[:len(p)])
	return len(p), nil
}

type registrationTestHarness struct {
	ctx       context.Context
	pool      *pgxpool.Pool
	container testcontainers.Container
	authRepo  auth.Repository
	userRepo  user.Repository
}

func newRegistrationTestHarness(t *testing.T) *registrationTestHarness {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "qq_registration_test",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(90 * time.Second),
		AutoRemove:   true,
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		if strings.Contains(err.Error(), "docker") {
			t.Skipf("skipping registration integration tests: %v", err)
		}
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to resolve container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to resolve container port: %v", err)
	}

	dsn := fmt.Sprintf(
		"postgres://postgres:postgres@%s/qq_registration_test?sslmode=disable", net.JoinHostPort(host, port.Port()))

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to create pgx pool: %v", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		_ = container.Terminate(ctx)
		t.Fatalf("database not reachable: %v", err)
	}

	applyRegistrationMigrations(t, context.Background(), pool)

	harness := &registrationTestHarness{
		ctx:       context.Background(),
		pool:      pool,
		container: container,
		authRepo:  auth.NewPgxRepository(pool),
		userRepo:  user.NewPgxRepository(pool),
	}

	t.Cleanup(func() {
		harness.Close()
	})

	return harness
}

func (h *registrationTestHarness) Close() {
	if h.pool != nil {
		h.pool.Close()
		h.pool = nil
	}
	if h.container != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = h.container.Terminate(ctx)
		h.container = nil
	}
}

func applyRegistrationMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	root := registrationProjectRoot(t)
	migrationsDir := filepath.Join(root, "db", "migrations")
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	require.NoError(t, err)
	sort.Strings(files)

	for _, file := range files {
		contents, err := os.ReadFile(file)
		require.NoErrorf(t, err, "failed to read migration %s", file)
		statements := splitSQLStatements(string(contents))
		for _, stmt := range statements {
			if stmt == "" {
				continue
			}
			_, err := pool.Exec(ctx, stmt)
			require.NoErrorf(t, err, "failed executing migration %s", file)
		}
	}
}

func splitSQLStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}
	return statements
}

func registrationProjectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("cannot determine caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}

func hashOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func createAuthAndUser(t *testing.T, h *registrationTestHarness, email, username string) (pgtype.UUID, db.User) {
	t.Helper()

	authIDPtr, err := h.authRepo.CreateAuthForOTPLogin(h.ctx, email)
	require.NoError(t, err)
	authID := *authIDPtr

	userRecord, err := h.userRepo.CreateUserWithAuthID(h.ctx, authID, username)
	require.NoError(t, err)

	return authID, *userRecord
}

func fetchUserByEmail(t *testing.T, svc user.Service, ctx context.Context, email string) *db.User {
	t.Helper()
	userRecord, err := svc.GetUserByEmail(ctx, email)
	require.NoError(t, err)
	return userRecord
}

func useDeterministicRand(t *testing.T, data []byte) {
	t.Helper()
	original := rand.Reader
	rand.Reader = &deterministicReader{data: data}
	t.Cleanup(func() {
		rand.Reader = original
	})
}

func countOTPs(t *testing.T, pool *pgxpool.Pool, authID pgtype.UUID) int {
	t.Helper()
	var count int
	require.NoError(
		t, pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM auth_otp_codes WHERE auth_id = $1", authID).Scan(&count),
	)
	return count
}
