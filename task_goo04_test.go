package goose

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestTaskGOO04 验证：迁移目录为空、或配置的迁移路径是普通文件（不是目录）时，
// goose 必须返回明确错误，而不能把这种情况当作"无需迁移"而静默成功。
func TestTaskGOO04(t *testing.T) {
	t.Run("empty_dir_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		migrations, err := CollectMigrations(dir, minVersion, maxVersion)
		if err == nil {
			t.Fatalf("expected error for empty migration dir %q, got nil error (migrations=%d): empty dir was treated as 'no migration needed'", dir, len(migrations))
		}
		if !strings.Contains(err.Error(), "no migration files found") {
			t.Fatalf("expected error containing %q, got: %v", "no migration files found", err)
		}
	})

	t.Run("not_a_dir_returns_error", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "not_a_dir")
		if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		migrations, err := CollectMigrations(filePath, minVersion, maxVersion)
		if err == nil {
			t.Fatalf("expected error when migration path %q is a regular file, got nil error (migrations=%d): non-directory path was treated as 'no migration needed'", filePath, len(migrations))
		}
	})
}
