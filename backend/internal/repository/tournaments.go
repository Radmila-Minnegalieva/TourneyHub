package repository

import (
	"context"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
)

// RosterRules содержит сохранённые в турнире ограничения дисциплины.
type RosterRules struct {
	Min  int
	Max  int
	Draw bool
	Unit string
}

// ApplicationFilter описывает выборку заявок без передачи SQL за пределы репозитория.
type ApplicationFilter struct {
	Limit        int
	Offset       int
	TournamentID *uuid.UUID
	Status       *model.ApplicationStatus
	UserID       *uuid.UUID
}

// Tournaments хранит турниры, заявки и снимки составов.
// Блокирующие методы вызываются внутри Store.WithinTx.
type Tournaments interface {
	ReadTournament(context.Context, uuid.UUID, bool) (model.Tournament, error)
	TournamentVisible(context.Context, uuid.UUID) (bool, error)
	ListTournamentPage(context.Context, model.ListTournamentsParams, *uuid.UUID, bool) (model.TournamentPage, error)
	ReadDisciplineRosterRules(context.Context, uuid.UUID) (RosterRules, error)
	ReadTournamentRosterRules(context.Context, uuid.UUID) (RosterRules, error)
	CreateTournamentRecord(context.Context, uuid.UUID, uuid.UUID, model.CreateTournamentRequest, RosterRules) error
	UpdateTournamentRecord(context.Context, uuid.UUID, model.CreateTournamentRequest, RosterRules) error
	TournamentRosterCompatible(context.Context, uuid.UUID, int, int) (bool, error)
	BumpTournamentVersion(context.Context, uuid.UUID) error
	SetTournamentStatus(context.Context, uuid.UUID, model.TournamentStatus) error
	SetTournamentReferee(context.Context, uuid.UUID, uuid.UUID) error
	TournamentPlayerIDs(context.Context, uuid.UUID) ([]uuid.UUID, error)
	MatchesUsingTournamentReferee(context.Context, uuid.UUID) ([]model.Match, error)
	ReadApplication(context.Context, uuid.UUID) (model.Application, error)
	LockTournamentApplication(context.Context, uuid.UUID, uuid.UUID) error
	LockApplicationTeam(context.Context, uuid.UUID) error
	ListApplicationPage(context.Context, ApplicationFilter) (model.ApplicationPage, error)
	ApplicationTeamCaptain(context.Context, uuid.UUID) (uuid.UUID, error)
	EligibleApplicationPlayersCount(context.Context, uuid.UUID, []uuid.UUID) (int, error)
	CreateApplicationRecord(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) error
	ReserveApplicationPlayers(context.Context, uuid.UUID, uuid.UUID, []uuid.UUID) error
	ReleaseApplicationPlayers(context.Context, uuid.UUID) error
	SetApplicationDecision(context.Context, uuid.UUID, model.ApplicationStatus, *string) error
	BumpApplicationVersion(context.Context, uuid.UUID) error
}
