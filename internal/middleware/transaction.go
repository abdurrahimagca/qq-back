package middleware

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txContextKey string

const TxContextKey txContextKey = "transaction"

// responseWriter wrapper to capture status code for transaction logic
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// TransactionMiddleware manages the lifecycle of a database transaction.
func TransactionMiddleware(db *pgxpool.Pool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tx, err := db.Begin(r.Context())
			if err != nil {
				http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
				return
			}

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			defer func() {
				if p := recover(); p != nil {
					tx.Rollback(r.Context())
					panic(p) // re-panic after rollback
				}
			}()

			ctx := context.WithValue(r.Context(), TxContextKey, tx)
			next.ServeHTTP(rw, r.WithContext(ctx))

			if rw.statusCode >= 200 && rw.statusCode < 300 {
				if err := tx.Commit(r.Context()); err != nil {
					// Log commit error, can't write to header as it's likely sent
				}
			} else {
				tx.Rollback(r.Context())
			}
		})
	}
}

// GetTxFromContext is a helper to retrieve the transaction from the context.
func GetTxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(TxContextKey).(pgx.Tx)
	return tx, ok
}
