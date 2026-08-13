package sqlparser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestTaskGOO02 verifies that the SQL parser does NOT trim content located
// after the last semicolon of a statement. The parser used to cut the input at
// the last semicolon (strings.LastIndex + truncate), which corrupted the tail
// of the final statement: trailing comments after the final semicolon, or the
// final statement line inside a StatementBegin/StatementEnd block that has no
// trailing semicolon, were silently dropped.
func TestTaskGOO02(t *testing.T) {
	t.Run("trailing comment after final semicolon preserved", func(t *testing.T) {
		input := `-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION task_goo02()
RETURNS trigger AS $$
BEGIN
    PERFORM 1;
END;
$$ LANGUAGE plpgsql; -- trailing comment must be preserved
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS task_goo02();
`
		statements, _, err := ParseSQLMigration(strings.NewReader(input), DirectionUp, false)
		require.NoError(t, err)
		require.Len(t, statements, 1)
		require.Contains(t, statements[0], "-- trailing comment must be preserved")
	})

	t.Run("final statement line without trailing semicolon preserved", func(t *testing.T) {
		input := `-- +goose Up
-- +goose StatementBegin
create table task_goo02 ( id int );
insert into task_goo02 values (1) -- no trailing semicolon but valid inside StatementBegin/End
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS task_goo02;
`
		statements, _, err := ParseSQLMigration(strings.NewReader(input), DirectionUp, false)
		require.NoError(t, err)
		require.Len(t, statements, 1)
		require.Contains(t, statements[0], "insert into task_goo02 values (1)")
	})
}
