package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/authctx"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

// Невостребованные методы Store намеренно не реализованы: сценарий сервиса
// проверяется через небольшое число типизированных операций без PostgreSQL.
type tournamentStoreRecorder struct {
	repository.Store
	tournament    model.Tournament
	transactions  int
	writes        int
	audits        int
	events        int
	inTransaction bool
}

func (r *tournamentStoreRecorder) WithinTx(ctx context.Context, fn func(repository.Store) error) error {
	r.transactions++
	r.inTransaction = true
	defer func() { r.inTransaction = false }()
	return fn(r)
}

func (r *tournamentStoreRecorder) ReadTournament(_ context.Context, _ uuid.UUID, _ bool) (model.Tournament, error) {
	return r.tournament, nil
}

func (r *tournamentStoreRecorder) SetTournamentStatus(_ context.Context, _ uuid.UUID, status model.TournamentStatus) error {
	if !r.inTransaction {
		return errors.New("status written outside transaction")
	}
	r.writes++
	r.tournament.Status = status
	return nil
}

func (r *tournamentStoreRecorder) BumpTournamentVersion(context.Context, uuid.UUID) error {
	if !r.inTransaction {
		return errors.New("version written outside transaction")
	}
	r.tournament.Version++
	return nil
}

func (r *tournamentStoreRecorder) Audit(context.Context, uuid.UUID, string, string, uuid.UUID, any, any) error {
	if !r.inTransaction {
		return errors.New("audit written outside transaction")
	}
	r.audits++
	return nil
}

func (r *tournamentStoreRecorder) Event(context.Context, uuid.UUID, string, any) error {
	if !r.inTransaction {
		return errors.New("event written outside transaction")
	}
	r.events++
	return nil
}

func serviceDraft(owner uuid.UUID) model.Tournament {
	r := validTournamentRequest()
	opens := time.Now().UTC()
	closes, ends := opens.Add(time.Hour), opens.Add(2*time.Hour)
	return model.Tournament{Id: uuid.New(), Organizer: model.PublicUser{Id: owner},
		Name: r.Name, DisciplineId: uuid.New(), Format: r.Format, MaxParticipants: r.MaxParticipants,
		Rules: r.Rules, Timezone: r.Timezone, Status: model.TournamentStatusDraft, Version: 1,
		RegistrationOpensAt: &opens, RegistrationClosesAt: &closes, StartsAt: &closes, EndsAt: &ends}
}

func TestPublishTournamentServiceWithTypedRepository(t *testing.T) {
	owner := uuid.New()
	store := &tournamentStoreRecorder{tournament: serviceDraft(owner)}
	s := &Service{repo: store}
	ctx := authctx.WithPrincipal(context.Background(), model.Principal{ID: owner, Roles: []model.Role{model.Organizer}})
	result, err := s.PublishTournament(ctx, store.tournament.Id, model.VersionRequest{Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != model.TournamentStatusRegistration || result.Version != 2 || store.transactions != 1 || store.writes != 1 || store.audits != 1 || store.events != 1 {
		t.Fatalf("publication did not commit one coherent operation: result=%+v store=%+v", result, store)
	}
}

func TestPublishTournamentServiceRejectsBeforeWriting(t *testing.T) {
	owner := uuid.New()
	for _, tc := range []struct {
		name    string
		actor   uuid.UUID
		version int
		want    apperror.Kind
	}{
		{"stale version", owner, 9, apperror.Conflict},
		{"foreign owner", uuid.New(), 1, apperror.Forbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			store := &tournamentStoreRecorder{tournament: serviceDraft(owner)}
			s := &Service{repo: store}
			ctx := authctx.WithPrincipal(context.Background(), model.Principal{ID: tc.actor, Roles: []model.Role{model.Organizer}})
			_, err := s.PublishTournament(ctx, store.tournament.Id, model.VersionRequest{Version: tc.version})
			var applicationError *apperror.Error
			if !errors.As(err, &applicationError) || applicationError.Kind != tc.want {
				t.Fatalf("error=%v, want kind %s", err, tc.want)
			}
			if store.writes != 0 || store.audits != 0 || store.events != 0 || store.tournament.Status != model.TournamentStatusDraft {
				t.Fatal("rejected operation changed tournament, audit or events")
			}
		})
	}
}
