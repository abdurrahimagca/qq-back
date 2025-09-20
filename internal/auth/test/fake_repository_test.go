package auth_test

import (
	"context"
	"sync"

	"github.com/abdurrahimagca/qq-back/internal/auth"
	"github.com/abdurrahimagca/qq-back/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeRepositoryState struct {
	mu                      sync.Mutex
	emailsByAuthID          map[string]string
	otpsByHash              map[string]db.GetUserIdAndEmailByOtpCodeRow
	userIDByAuthID          map[string]pgtype.UUID
	killOrphanedEmails      []string
	killOrphanedUserIDs     []pgtype.UUID
	createAuthErr           error
	nextAuthID              *pgtype.UUID
	createOTPErr            error
	getOtpErr               error
	killOrphanedErr         error
	killOrphanedByUserIDErr error
	withTxCount             int
	lastTx                  pgx.Tx
}

type fakeRepository struct {
	state *fakeRepositoryState
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		state: &fakeRepositoryState{
			emailsByAuthID:      make(map[string]string),
			otpsByHash:          make(map[string]db.GetUserIdAndEmailByOtpCodeRow),
			userIDByAuthID:      make(map[string]pgtype.UUID),
			killOrphanedEmails:  make([]string, 0),
			killOrphanedUserIDs: make([]pgtype.UUID, 0),
		},
	}
}

func (f *fakeRepository) WithTx(tx pgx.Tx) auth.Repository {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.withTxCount++
	f.state.lastTx = tx
	return &fakeRepository{state: f.state}
}

func (f *fakeRepository) CreateAuthForOTPLogin(ctx context.Context, email string) (*pgtype.UUID, error) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()

	if f.state.createAuthErr != nil {
		return nil, f.state.createAuthErr
	}

	var id pgtype.UUID
	if f.state.nextAuthID != nil {
		id = *f.state.nextAuthID
		f.state.nextAuthID = nil
	} else {
		id = newPGUUID()
	}

	f.state.emailsByAuthID[uuidToString(id)] = email
	idCopy := id
	return &idCopy, nil
}

func (f *fakeRepository) CreateOTP(ctx context.Context, authID pgtype.UUID, otpHash string) error {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()

	if f.state.createOTPErr != nil {
		return f.state.createOTPErr
	}

	entry := db.GetUserIdAndEmailByOtpCodeRow{AuthID: authID}
	if email, ok := f.state.emailsByAuthID[uuidToString(authID)]; ok {
		entry.Email = email
	}
	if userID, ok := f.state.userIDByAuthID[uuidToString(authID)]; ok {
		entry.ID = userID
	}

	f.state.otpsByHash[otpHash] = entry
	return nil
}

func (f *fakeRepository) GetUserIDAndEmailByOTPCode(
	ctx context.Context, otpHash string) (db.GetUserIdAndEmailByOtpCodeRow, error) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()

	if f.state.getOtpErr != nil {
		return db.GetUserIdAndEmailByOtpCodeRow{}, f.state.getOtpErr
	}

	if entry, ok := f.state.otpsByHash[otpHash]; ok {
		return entry, nil
	}

	return db.GetUserIdAndEmailByOtpCodeRow{}, auth.ErrNotFound
}

func (f *fakeRepository) KillOrphanedOTPs(ctx context.Context, email string) error {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()

	f.state.killOrphanedEmails = append(f.state.killOrphanedEmails, email)

	if f.state.killOrphanedErr != nil {
		return f.state.killOrphanedErr
	}

	for hash, entry := range f.state.otpsByHash {
		if entry.Email == email {
			delete(f.state.otpsByHash, hash)
		}
	}

	return nil
}

func (f *fakeRepository) KillOrphanedOTPsByUserID(ctx context.Context, userID pgtype.UUID) error {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()

	f.state.killOrphanedUserIDs = append(f.state.killOrphanedUserIDs, userID)

	if f.state.killOrphanedByUserIDErr != nil {
		return f.state.killOrphanedByUserIDErr
	}

	for hash, entry := range f.state.otpsByHash {
		if entry.ID == userID {
			delete(f.state.otpsByHash, hash)
		}
	}

	return nil
}

// helper configuration methods

func (f *fakeRepository) setCreateAuthErr(err error) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.createAuthErr = err
}

func (f *fakeRepository) setNextAuthID(id pgtype.UUID) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.nextAuthID = &id
}

func (f *fakeRepository) setCreateOTPErr(err error) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.createOTPErr = err
}

func (f *fakeRepository) setGetOtpErr(err error) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.getOtpErr = err
}

func (f *fakeRepository) setKillOrphanedErr(err error) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.killOrphanedErr = err
}

func (f *fakeRepository) setKillOrphanedByUserIDErr(err error) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.killOrphanedByUserIDErr = err
}

func (f *fakeRepository) setAuthEmail(authID pgtype.UUID, email string) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.emailsByAuthID[uuidToString(authID)] = email
}

func (f *fakeRepository) setUserForAuth(authID, userID pgtype.UUID) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.userIDByAuthID[uuidToString(authID)] = userID
}

func (f *fakeRepository) setOTP(hash string, row db.GetUserIdAndEmailByOtpCodeRow) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	f.state.otpsByHash[hash] = row
}

func (f *fakeRepository) getOTP(hash string) (db.GetUserIdAndEmailByOtpCodeRow, bool) {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	row, ok := f.state.otpsByHash[hash]
	return row, ok
}

func (f *fakeRepository) withTxCount() int {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	return f.state.withTxCount
}

func (f *fakeRepository) killOrphanedEmails() []string {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	emails := make([]string, len(f.state.killOrphanedEmails))
	copy(emails, f.state.killOrphanedEmails)
	return emails
}

func (f *fakeRepository) killOrphanedUserIDs() []pgtype.UUID {
	f.state.mu.Lock()
	defer f.state.mu.Unlock()
	ids := make([]pgtype.UUID, len(f.state.killOrphanedUserIDs))
	copy(ids, f.state.killOrphanedUserIDs)
	return ids
}
