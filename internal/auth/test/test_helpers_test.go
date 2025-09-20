package auth_test

import (
	cript "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
		return len(d.data), io.EOF
	}
	copy(p, d.data[:len(p)])
	return len(p), nil
}

func useRandReader(t *testing.T, reader io.Reader) {
	t.Helper()
	original := cript.Reader
	cript.Reader = reader
	t.Cleanup(func() {
		cript.Reader = original
	})
}

func hashOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}

func newPGUUID() pgtype.UUID {
	u := uuid.New()
	var id pgtype.UUID
	copy(id.Bytes[:], u[:])
	id.Valid = true
	return id
}

func uuidToString(id pgtype.UUID) string {
	return uuid.UUID(id.Bytes).String()
}
