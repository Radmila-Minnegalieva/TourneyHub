package repository

import (
	"context"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
	"time"
)

type Ancillary interface {
	AdminCatalog
	Resources
}

type StoredFile struct {
	File    model.File
	OwnerID uuid.UUID
	Data    []byte
}

type ExportRow struct {
	Kind      string
	Place     string
	Team      string
	MatchID   string
	HomeTeam  string
	AwayTeam  string
	HomeScore string
	AwayScore string
	Status    string
}

type PendingMail struct {
	ID      int64
	Message model.MailMessage
}

type BootstrapState struct {
	AdminCount      int
	SameActiveAdmin bool
}

type Resources interface {
	CreateFile(context.Context, uuid.UUID, uuid.UUID, model.FileUpload, string) (model.File, error)
	File(context.Context, uuid.UUID) (StoredFile, error)
	CanReadProtocol(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	ExportRows(context.Context, uuid.UUID) ([]ExportRow, error)
	EventBounds(context.Context, uuid.UUID) (int64, int64, error)
	EventsAfter(context.Context, uuid.UUID, int64, int) ([]model.TournamentEvent, error)
	PendingMail(context.Context) (PendingMail, error)
	CompleteMail(context.Context, int64) error
	RetryMail(context.Context, int64) error
	CleanupRateLimits(context.Context) error
	CleanupEvents(context.Context, time.Time) error
	BootstrapState(context.Context, string) (BootstrapState, error)
	CreateBootstrapAdmin(context.Context, uuid.UUID, string, string, string) error
}
