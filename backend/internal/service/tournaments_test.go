package service

import (
	"context"
	"testing"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/authctx"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"

	"github.com/google/uuid"
)

func validTournamentRequest() model.CreateTournamentRequest {
	return model.CreateTournamentRequest{
		Name: "University Cup", Format: model.TournamentFormatSingleElimination,
		MaxParticipants: 8, Timezone: "Europe/Moscow",
		Rules: model.Rules{BestOf: 1, RoundRobinLegs: 1, WinPoints: 3, DrawPoints: 1,
			TieBreakers: []model.TieBreaker{model.TieBreakerScoreDifference}, ConfirmationTimeoutHours: 24},
	}
}

func TestTournamentPublicationDates(t *testing.T) {
	base := time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)
	date := func(offset time.Duration) *time.Time { value := base.Add(offset); return &value }
	for _, tc := range []struct {
		name      string
		opens     *time.Time
		closes    *time.Time
		starts    *time.Time
		ends      *time.Time
		publish   bool
		wantError bool
	}{
		{name: "undated draft"},
		{name: "undated publication", publish: true, wantError: true},
		{name: "valid publication", opens: date(0), closes: date(time.Hour), starts: date(time.Hour), ends: date(2 * time.Hour), publish: true},
		{name: "closed before opening", opens: date(time.Hour), closes: date(0), wantError: true},
		{name: "equal opening and closing", opens: date(0), closes: date(0), wantError: true},
		{name: "starts before registration closes", closes: date(time.Hour), starts: date(0), wantError: true},
		{name: "zero tournament duration", starts: date(0), ends: date(0), wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := validTournamentRequest()
			r.RegistrationOpensAt, r.RegistrationClosesAt, r.StartsAt, r.EndsAt = tc.opens, tc.closes, tc.starts, tc.ends
			if err := validateTournamentRequest(r, tc.publish); (err != nil) != tc.wantError {
				t.Fatalf("validation error=%v, wantError=%v", err, tc.wantError)
			}
		})
	}
}

func TestTournamentRulesRejectInconsistentGroups(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*model.CreateTournamentRequest)
	}{
		{"missing group settings", func(r *model.CreateTournamentRequest) { r.Format = model.TournamentFormatGroupsPlayoffs }},
		{"groups on elimination", func(r *model.CreateTournamentRequest) {
			r.Rules.GroupSettings = &model.GroupSettings{GroupCount: 2, QualifiersPerGroup: 1}
		}},
		{"too many groups", func(r *model.CreateTournamentRequest) {
			r.Format = model.TournamentFormatGroupsPlayoffs
			r.Rules.GroupSettings = &model.GroupSettings{GroupCount: 5, QualifiersPerGroup: 1}
		}},
		{"wildcard without place", func(r *model.CreateTournamentRequest) {
			r.Format = model.TournamentFormatGroupsPlayoffs
			r.Rules.GroupSettings = &model.GroupSettings{GroupCount: 2, QualifiersPerGroup: 1, WildcardCount: 1}
		}},
		{"duplicate tie breaker", func(r *model.CreateTournamentRequest) {
			r.Rules.TieBreakers = append(r.Rules.TieBreakers, r.Rules.TieBreakers[0])
		}},
		{"invalid timezone", func(r *model.CreateTournamentRequest) { r.Timezone = "Mars/Unknown" }},
		{"empty timezone", func(r *model.CreateTournamentRequest) { r.Timezone = "" }},
		{"host timezone", func(r *model.CreateTournamentRequest) { r.Timezone = "Local" }},
		{"round robin third place", func(r *model.CreateTournamentRequest) {
			r.Format = model.TournamentFormatRoundRobin
			r.Rules.ThirdPlaceMatch = true
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := validTournamentRequest()
			tc.mutate(&r)
			if err := validateTournamentRequest(r, false); err == nil {
				t.Fatal("inconsistent rules accepted")
			}
		})
	}
	place := 2
	r := validTournamentRequest()
	r.Format = model.TournamentFormatGroupsPlayoffs
	r.Rules.GroupSettings = &model.GroupSettings{GroupCount: 2, QualifiersPerGroup: 1, WildcardCount: 1, WildcardPlace: &place}
	if err := validateTournamentRequest(r, false); err != nil {
		t.Fatalf("valid groups rejected: %v", err)
	}
}

func TestTournamentPatchKeepsOmittedFields(t *testing.T) {
	r := validTournamentRequest()
	tournament := model.Tournament{Name: r.Name, Format: r.Format, MaxParticipants: r.MaxParticipants, Timezone: r.Timezone, Rules: r.Rules}
	name := "Renamed cup"
	merged := mergeTournament(tournament, model.UpdateTournamentRequest{Name: &name, Version: 1})
	if merged.Name != name || merged.Format != r.Format || merged.Rules.BestOf != r.Rules.BestOf || merged.MaxParticipants != r.MaxParticipants {
		t.Fatalf("patch changed omitted fields: %+v", merged)
	}
}

func TestRegistrationBoundary(t *testing.T) {
	start := time.Now().UTC()
	end := start.Add(time.Hour)
	tournament := model.Tournament{Status: model.TournamentStatusRegistration, RegistrationOpensAt: &start, RegistrationClosesAt: &end}
	if registrationOpen(tournament, start.Add(-time.Nanosecond)) || !registrationOpen(tournament, start) || registrationOpen(tournament, end) {
		t.Fatal("registration must include opening time and exclude closing time")
	}
	tournament.Hidden = true
	if registrationOpen(tournament, start) {
		t.Fatal("moderated tournament accepted an application")
	}
}

func TestTournamentOwnership(t *testing.T) {
	owner, stranger := uuid.New(), uuid.New()
	tournament := model.Tournament{Organizer: model.PublicUser{Id: owner}}
	for _, tc := range []struct {
		name    string
		id      uuid.UUID
		role    model.Role
		allowed bool
	}{
		{"owner organizer", owner, model.Organizer, true},
		{"foreign organizer", stranger, model.Organizer, false},
		{"owner without organizer role", owner, model.Participant, false},
		{"administrator", stranger, model.Admin, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := authctx.WithPrincipal(context.Background(), model.Principal{ID: tc.id, Roles: []model.Role{tc.role}})
			if err := requireTournamentOwner(ctx, tournament); (err == nil) != tc.allowed {
				t.Fatalf("ownership error=%v, allowed=%v", err, tc.allowed)
			}
		})
	}
}
