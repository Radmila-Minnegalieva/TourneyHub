package repository

import (
	"context"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
)

// AdminCatalog описывает хранение справочников, рейтингов и административных данных.
type AdminCatalog interface {
	AdminUserPage(context.Context, model.ListAdminUsersParams) (model.ProfilePage, error)
	UserHasAdminRole(context.Context, uuid.UUID) (bool, error)
	OtherActiveAdminCount(context.Context, uuid.UUID) (int, error)
	LockUserVersion(context.Context, uuid.UUID) (int, error)
	SetUserBlocked(context.Context, uuid.UUID, bool, *string) error
	ReplaceUserRoles(context.Context, uuid.UUID, []model.Role) error
	SetTournamentVisibility(context.Context, uuid.UUID, bool) error
	SoftDeleteTournamentRecord(context.Context, uuid.UUID) error
	AuditPage(context.Context, model.ListAuditParams) (model.AuditEntryPage, error)
	DisciplinePage(context.Context, int, int) (model.DisciplinePage, error)
	CreateDisciplineRecord(context.Context, model.DisciplineRequest) (model.Discipline, error)
	Discipline(context.Context, uuid.UUID, bool) (model.Discipline, error)
	UpdateDisciplineRecord(context.Context, uuid.UUID, model.Discipline) (model.Discipline, error)
	RatingPage(context.Context, uuid.UUID, int, int) (model.RatingPage, error)
	RatingHistoryPage(context.Context, uuid.UUID, uuid.UUID, int, int) (model.RatingChangePage, error)
	NotificationPage(context.Context, uuid.UUID, *bool, int, int) (model.NotificationPage, error)
	MarkNotificationRead(context.Context, uuid.UUID, uuid.UUID) (bool, error)
}
