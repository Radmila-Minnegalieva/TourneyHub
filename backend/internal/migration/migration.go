// Package migration применяет SQL-миграции goose из директории на диске.
package migration

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

// Open проверяет директорию, подключается к PostgreSQL и создаёт provider goose.
// Блокировка PostgreSQL защищает от одновременного запуска миграций.
// Вызывающий код должен закрыть возвращённое соединение с БД.
func Open(ctx context.Context, dsn, directory string) (*goose.Provider, *sql.DB, error) {
	info, err := os.Stat(directory)
	if err != nil {
		return nil, nil, fmt.Errorf("открыть директорию миграций %s: %w", directory, err)
	}
	if !info.IsDir() {
		return nil, nil, fmt.Errorf("путь миграций %s должен указывать на директорию", directory)
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("открыть БД для миграций: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("проверить БД для миграций: %w", err)
	}
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(directory), goose.WithSessionLocker(locker))
	if err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("создать provider миграций: %w", err)
	}
	return provider, db, nil
}
