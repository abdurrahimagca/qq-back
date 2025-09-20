package auth_test

import (
	"context"
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
	qqerrors "github.com/abdurrahimagca/qq-back/internal/utils/errors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type integrationHarness struct {
	container testcontainers.Container
	pool      *pgxpool.Pool
	repo      auth.Repository
	queries   *db.Queries
}

func setupIntegrationHarness(t *testing.T) *integrationHarness {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image: "postgres:16-alpine",
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "qq_db_test",
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
			t.Skipf("skipping integration tests: %v", err)
		}
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to resolve container host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to resolve container port: %v", err)
	}

	dsn := fmt.Sprintf(
		"postgres://postgres:postgres@%s/qq_db_test?sslmode=disable", net.JoinHostPort(host, mappedPort.Port()))

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

	applyMigrations(t, context.Background(), pool)

	harness := &integrationHarness{
		container: container,
		pool:      pool,
		repo:      auth.NewPgxRepository(pool),
		queries:   db.New(pool),
	}

	t.Cleanup(func() {
		harness.Close()
	})

	return harness
}

func (h *integrationHarness) Close() {
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

func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	root := projectRoot(t)
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
	raw := strings.Split(sql, ";")
	statements := make([]string, 0, len(raw))
	for _, stmt := range raw {
		trimmed := strings.TrimSpace(stmt)
		if trimmed != "" {
			statements = append(statements, trimmed)
		}
	}
	return statements
}

func projectRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("cannot determine caller path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}

func (h *integrationHarness) createUserForAuth(ctx context.Context, authID pgtype.UUID) (pgtype.UUID, error) {
	params := db.InsertUserParams{
		AuthID:      authID,
		Username:    fmt.Sprintf("user_%d", time.Now().UnixNano()),
		DisplayName: pgtype.Text{Valid: false},
		AvatarKey:   pgtype.Text{Valid: false},
	}
	user, err := h.queries.InsertUser(ctx, params)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return user.ID, nil
}

func countRows(ctx context.Context, pool *pgxpool.Pool, query string, args ...any) (int, error) {
	var count int
	err := pool.QueryRow(ctx, query, args...).Scan(&count)
	return count, err
}

func TestPgxRepository_CreateAuthForOTPLogin(t *testing.T) {
	h := setupIntegrationHarness(t)

	ctx := context.Background()
	email := fmt.Sprintf("auth-%d@example.com", time.Now().UnixNano())

	id, err := h.repo.CreateAuthForOTPLogin(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, id)

	var provider string
	err = h.pool.QueryRow(ctx, "SELECT provider FROM auth WHERE id = $1", *id).Scan(&provider)
	require.NoError(t, err)
	require.Equal(t, "email_otp", provider)

	_, err = h.repo.CreateAuthForOTPLogin(ctx, email)
	require.Error(t, err)
	var qqErr *qqerrors.QQError
	require.ErrorAs(t, err, &qqErr)
	require.Equal(t, qqerrors.ErrUniqueViolation, qqErr.Original)
}

func TestPgxRepository_CreateOTP(t *testing.T) {
	h := setupIntegrationHarness(t)

	ctx := context.Background()
	email := fmt.Sprintf("otp-%d@example.com", time.Now().UnixNano())
	authID, err := h.repo.CreateAuthForOTPLogin(ctx, email)
	require.NoError(t, err)

	hash := hashOTP("ABC123")
	err = h.repo.CreateOTP(ctx, *authID, hash)
	require.NoError(t, err)

	var stored string
	err = h.pool.QueryRow(ctx, "SELECT code FROM auth_otp_codes WHERE auth_id = $1", *authID).Scan(&stored)
	require.NoError(t, err)
	require.Equal(t, hash, stored)
}

func TestPgxRepository_GetUserIDAndEmailByOTPCode(t *testing.T) {
	h := setupIntegrationHarness(t)

	ctx := context.Background()
	email := fmt.Sprintf("user-%d@example.com", time.Now().UnixNano())
	authID, err := h.repo.CreateAuthForOTPLogin(ctx, email)
	require.NoError(t, err)

	userID, err := h.createUserForAuth(ctx, *authID)
	require.NoError(t, err)

	otpCode := "ABC123"
	hash := hashOTP(otpCode)
	err = h.repo.CreateOTP(ctx, *authID, hash)
	require.NoError(t, err)

	row, err := h.repo.GetUserIDAndEmailByOTPCode(ctx, hash)
	require.NoError(t, err)
	require.Equal(t, email, row.Email)
	require.Equal(t, uuidToString(*authID), uuidToString(row.AuthID))
	require.Equal(t, uuidToString(userID), uuidToString(row.ID))

	_, err = h.repo.GetUserIDAndEmailByOTPCode(ctx, "missing")
	require.ErrorIs(t, err, auth.ErrNotFound)
}

func TestPgxRepository_KillOrphanedOTPs(t *testing.T) {
	h := setupIntegrationHarness(t)

	ctx := context.Background()
	email := fmt.Sprintf("cleanup-%d@example.com", time.Now().UnixNano())
	authID, err := h.repo.CreateAuthForOTPLogin(ctx, email)
	require.NoError(t, err)
	err = h.repo.CreateOTP(ctx, *authID, hashOTP("ABC123"))
	require.NoError(t, err)

	err = h.repo.KillOrphanedOTPs(ctx, email)
	require.NoError(t, err)

	count, err := countRows(ctx, h.pool, "SELECT COUNT(*) FROM auth_otp_codes WHERE auth_id = $1", *authID)
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestPgxRepository_KillOrphanedOTPsByUserID(t *testing.T) {
	h := setupIntegrationHarness(t)

	ctx := context.Background()
	email := fmt.Sprintf("cleanup-user-%d@example.com", time.Now().UnixNano())
	authID, err := h.repo.CreateAuthForOTPLogin(ctx, email)
	require.NoError(t, err)
	userID, err := h.createUserForAuth(ctx, *authID)
	require.NoError(t, err)

	err = h.repo.CreateOTP(ctx, *authID, hashOTP("ABC123"))
	require.NoError(t, err)

	err = h.repo.KillOrphanedOTPsByUserID(ctx, userID)
	require.NoError(t, err)

	count, err := countRows(ctx, h.pool, "SELECT COUNT(*) FROM auth_otp_codes WHERE auth_id = $1", *authID)
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestPgxRepository_WithTx_Rollback(t *testing.T) {
	h := setupIntegrationHarness(t)

	ctx := context.Background()
	tx, err := h.pool.Begin(ctx)
	require.NoError(t, err)

	repoTx := h.repo.WithTx(tx)
	email := fmt.Sprintf("tx-%d@example.com", time.Now().UnixNano())
	_, err = repoTx.CreateAuthForOTPLogin(ctx, email)
	require.NoError(t, err)

	err = tx.Rollback(ctx)
	require.NoError(t, err)

	count, err := countRows(ctx, h.pool, "SELECT COUNT(*) FROM auth WHERE email = $1", email)
	require.NoError(t, err)
	require.Zero(t, count)
}

func TestPgxRepository_GetUserIDAndEmailByOTPCode_DBError(t *testing.T) {
	h := setupIntegrationHarness(t)

	h.pool.Close()
	h.pool = nil

	_, err := h.repo.GetUserIDAndEmailByOTPCode(context.Background(), "deadbeef")
	require.Error(t, err)
	var qqErr *qqerrors.QQError
	require.ErrorAs(t, err, &qqErr)
}

func TestPgxRepository_KillOrphanedOTPs_DBError(t *testing.T) {
	h := setupIntegrationHarness(t)

	h.pool.Close()
	h.pool = nil

	err := h.repo.KillOrphanedOTPs(context.Background(), "test@example.com")
	require.Error(t, err)
	var qqErr *qqerrors.QQError
	require.ErrorAs(t, err, &qqErr)
}
