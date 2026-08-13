package goose_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/internal/check"
)

// TestTaskGOO01GoMigrationPanicIsReturnedAsError reproduces the target defect: a
// user-registered Go migration that panics must not crash the process. The migration
// framework must recover the panic and return it as an error.
//
// In the broken (Before) state this test fails because the panic escapes out of the
// migration framework and is never surfaced as an error. In the fixed (After) state the
// panic is converted into an error that contains the panic message and the process
// keeps running (proven by the assertions that execute after p.Up returns).
func TestTaskGOO01GoMigrationPanicIsReturnedAsError(t *testing.T) {
	ctx := context.Background()
	const wantErrString = "panic: runtime error: index out of range [7] with length 0"

	migration := goose.NewGoMigration(
		1,
		&goose.GoFunc{
			RunTx: func(ctx context.Context, tx *sql.Tx) error {
				var ss []int
				_ = ss[7] // deterministic panic: index out of range
				return nil
			},
		},
		nil,
	)
	p, err := goose.NewProvider(goose.DialectSQLite3, newDB(t), nil,
		goose.WithGoMigrations(migration),
	)
	check.NoError(t, err)

	// If the framework does not recover the panic, this call panics and the test fails.
	_, err = p.Up(ctx)

	// The panic must be surfaced as an error that contains the panic message; reaching
	// these assertions proves the process survived the panicking migration.
	check.HasError(t, err)
	check.Contains(t, err.Error(), wantErrString)
}
