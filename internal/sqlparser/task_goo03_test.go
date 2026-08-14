package sqlparser

import (
	"strings"
	"testing"
)

// TestTaskGOO03UnterminatedUpStatementIsError 针对固定任务卡 GOO03 的回归测试：
// Up 区段最后一条 SQL 没有分号终止符、且后续出现 "-- +goose Down" 注解时，
// 解析器必须明确报错，而不是静默忽略该语句并返回成功（迁移显示成功但实际少执行一条语句）。
func TestTaskGOO03UnterminatedUpStatementIsError(t *testing.T) {
	t.Parallel()

	input := "-- +goose Up\n" +
		"CREATE TABLE task_goo03 (id INTEGER PRIMARY KEY);\n" +
		"INSERT INTO task_goo03 (id) VALUES (1)\n" + // 最后一条 up 语句没有终止符
		"-- +goose Down\n" +
		"DROP TABLE task_goo03;\n"

	stmts, _, err := ParseSQLMigration(strings.NewReader(input), DirectionUp, false)
	if err == nil {
		t.Fatalf("expected an error for the unterminated last up statement, got nil; "+
			"parsed %d up statement(s): the last statement was silently ignored and the migration reports success", len(stmts))
	}
}
