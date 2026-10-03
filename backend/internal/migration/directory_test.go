package migration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/migration"
)

func TestOpenRejectsInvalidDirectoryBeforeConnecting(t *testing.T) {
	file := filepath.Join(t.TempDir(), "file.sql")
	if err := os.WriteFile(file, nil, 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{file, filepath.Join(t.TempDir(), "missing")} {
		provider, db, err := migration.Open(context.Background(), "invalid DSN", path)
		if err == nil || provider != nil || db != nil {
			t.Fatal("неверный путь должен отклоняться до подключения к БД")
		}
	}
}
