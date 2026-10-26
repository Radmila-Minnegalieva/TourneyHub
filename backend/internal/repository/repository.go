// Package repository описывает операции хранения данных, используемые сервисами.
package repository

import (
	"context"
	"errors"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("repository: record not found")

// Store объединяет репозитории в одной сессии хранения. Внутри WithinTx все
// операции переданного Store используют одну транзакцию.
type Store interface {
	Identity
	Teams
	Tournaments
	Competition
	Ancillary
	Common
	WithinTx(context.Context, func(Store) error) error
}

type NotificationRecipient struct {
	Email   string
	Enabled bool
}

// Common содержит общие записи аудита, outbox и блокировки бизнес-операций.
type Common interface {
	Audit(context.Context, uuid.UUID, string, string, uuid.UUID, any, any) error
	Outbox(context.Context, string, uuid.UUID, any) error
	Notify(context.Context, uuid.UUID, string, string, uuid.UUID) error
	NotificationRecipient(context.Context, uuid.UUID) (NotificationRecipient, error)
	Event(context.Context, uuid.UUID, string, any) error
	LockSchedule(context.Context) error
	LockRatings(context.Context, uuid.UUID) error
	LockAdmin(context.Context) error
}
