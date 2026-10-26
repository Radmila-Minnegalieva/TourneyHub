package postgres

import (
	"context"
	"encoding/json"
)

// jsonItems материализует SQL JSON-представление и закрывает курсор до выхода из репозитория.
func jsonItems[T any](ctx context.Context, q DBTX, query string, args ...any) ([]T, error) {
	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	items := make([]T, 0)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, dbError(err)
		}
		var item T
		if err := json.Unmarshal(raw, &item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, dbError(rows.Err())
}

func jsonObject[T any](ctx context.Context, q DBTX, query string, args ...any) (T, error) {
	var item T
	var raw []byte
	if err := q.QueryRow(ctx, query, args...).Scan(&raw); err != nil {
		return item, dbError(err)
	}
	err := json.Unmarshal(raw, &item)
	return item, err
}
