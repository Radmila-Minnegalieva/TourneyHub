package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/authctx"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

// В проверках отказа достаточно узких операций чтения: попытка записать результат
// через неиспользуемую часть встроенного Store немедленно обнаружится тестом.
type confirmationStore struct {
	repository.Store
	match        model.Match
	tournament   model.Tournament
	report       model.ResultReport
	captain      uuid.UUID
	side         int
	transactions int
}

func (q *confirmationStore) MatchTournamentID(context.Context, uuid.UUID) (uuid.UUID, error) {
	return q.tournament.Id, nil
}
func (q *confirmationStore) WithinTx(ctx context.Context, fn func(repository.Store) error) error {
	q.transactions++
	return fn(q)
}
func (q *confirmationStore) ReadTournament(context.Context, uuid.UUID, bool) (model.Tournament, error) {
	return q.tournament, nil
}
func (q *confirmationStore) ReadMatch(context.Context, uuid.UUID, bool) (model.Match, error) {
	return q.match, nil
}
func (q *confirmationStore) ReadResultReport(context.Context, uuid.UUID, uuid.UUID) (model.ResultReport, error) {
	return q.report, nil
}
func (q *confirmationStore) MatchCaptainSide(context.Context, *uuid.UUID, *uuid.UUID, uuid.UUID) (int, error) {
	return q.side, nil
}
func (q *confirmationStore) ReadReportingApplication(context.Context, uuid.UUID) (*uuid.UUID, error) {
	return q.match.HomeApplicationId, nil
}

func TestConfirmationRejectsWrongSideAndStaleVersionWithoutStorageWrites(t *testing.T) {
	home, away, captain := uuid.New(), uuid.New(), uuid.New()
	for _, tc := range []struct {
		name    string
		side    int
		version int
		kind    apperror.Kind
	}{
		{"reporting side", 1, 2, apperror.Forbidden},
		{"opposing side with stale version", 2, 1, apperror.Conflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			q := &confirmationStore{match: model.Match{Id: uuid.New(), HomeApplicationId: &home, AwayApplicationId: &away, Version: 2, Status: model.MatchStatusAwaitingConfirmation}, tournament: model.Tournament{Id: uuid.New(), Status: model.TournamentStatusInProgress}, report: model.ResultReport{Id: uuid.New(), AuthorId: uuid.New(), Status: model.ResultReportStatusPending}, captain: captain, side: tc.side}
			ctx := authctx.WithPrincipal(context.Background(), model.Principal{ID: captain, Roles: []model.Role{model.Participant}})
			_, err := New(q, config.Default(), nil).ConfirmResult(ctx, q.match.Id, q.report.Id, model.VersionRequest{Version: tc.version})
			var domainErr *apperror.Error
			if !errors.As(err, &domainErr) || domainErr.Kind != tc.kind {
				t.Fatalf("error=%v, want=%s", err, tc.kind)
			}
			if q.transactions != 1 {
				t.Fatalf("transactions=%d, want=1", q.transactions)
			}
		})
	}
}
