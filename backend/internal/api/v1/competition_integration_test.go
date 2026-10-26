//go:build integration

package api

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository/postgres"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/service"
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type competitionFixture struct {
	server       *echo.Echo
	db           *pgxpool.Pool
	tournament   uuid.UUID
	discipline   uuid.UUID
	organizer    contract.Tokens
	referee      contract.Tokens
	applications []uuid.UUID
	captains     map[uuid.UUID]contract.Tokens
}

func newCompetitionFixture(t *testing.T, format contract.TournamentFormat, n int) competitionFixture {
	t.Helper()
	server, db := newIntegrationServer(t)
	ctx := context.Background()
	organizerID, organizer := seedIntegrationUser(t, db, "organizer", contract.Organizer)
	refereeID, referee := seedIntegrationUser(t, db, "referee", contract.Referee)
	f := competitionFixture{server: server, db: db, tournament: uuid.New(), discipline: uuid.New(), organizer: organizer, referee: referee, captains: map[uuid.UUID]contract.Tokens{}}
	_, err := db.Exec(ctx, `INSERT INTO disciplines(id,name,min_roster_size,max_roster_size,allows_draw,score_unit) VALUES($1,'Board game',1,1,true,'points')`, f.discipline)
	if err != nil {
		t.Fatal(err)
	}
	var groups, qualifiers *int
	if format == contract.TournamentFormatGroupsPlayoffs {
		g, q := 2, 1
		groups, qualifiers = &g, &q
	}
	_, err = db.Exec(ctx, `INSERT INTO tournaments(id,organizer_id,discipline_id,referee_id,name,format,status,max_participants,registration_opens_at,registration_closes_at,starts_at,ends_at,timezone,best_of,round_robin_legs,win_points,draw_points,loss_points,tie_breakers,group_count,qualifiers_per_group,min_roster_size,max_roster_size,allows_draw,score_unit) VALUES($1,$2,$3,$4,'Competition',$5,'registration',256,now()-interval '3 days',now()-interval '2 days',now()-interval '1 day',now()+interval '7 days','UTC',1,1,3,1,0,ARRAY['head_to_head','score_difference','score_for'],$6,$7,1,1,true,'points')`, f.tournament, organizerID, f.discipline, refereeID, format, groups, qualifiers)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		user, tokens := seedIntegrationUser(t, db, fmt.Sprintf("player%d", i))
		team, app := uuid.New(), uuid.New()
		tx, e := db.Begin(ctx)
		if e != nil {
			t.Fatal(e)
		}
		_, err = tx.Exec(ctx, `INSERT INTO teams(id,name,captain_id) VALUES($1,$2,$3)`, team, fmt.Sprintf("Team %d", i), user)
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, team, user)
		}
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO applications(id,tournament_id,team_id,submitted_by,status) VALUES($1,$2,$3,$4,'accepted')`, app, f.tournament, team, user)
		}
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO application_players(application_id,tournament_id,user_id) VALUES($1,$2,$3)`, app, f.tournament, user)
		}
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO tournament_player_claims(tournament_id,user_id,application_id) VALUES($1,$2,$3)`, f.tournament, user, app)
		}
		if err != nil {
			tx.Rollback(ctx)
			t.Fatal(err)
		}
		if err = tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		f.applications = append(f.applications, app)
		f.captains[app] = tokens
	}
	return f
}
func (f competitionFixture) draw(t *testing.T) contract.Bracket {
	t.Helper()
	path := BaseURL + "/tournaments/" + f.tournament.String()
	req := contract.SeedingRequest{Method: contract.SeedingRequestMethodManual, ApplicationIds: &f.applications, Version: 1}
	preview := integrationDecode[contract.DrawPreview](t, integrationRequest(t, f.server, "POST", path+"/draw-previews", stringMustJSON(t, req), f.organizer.AccessToken, 201))
	confirm := stringMustJSON(t, contract.ConfirmDrawRequest{PreviewId: preview.Id, Version: 1})
	bracket := integrationDecode[contract.Bracket](t, integrationRequest(t, f.server, "PUT", path+"/draw", confirm, f.organizer.AccessToken, 200))
	integrationRequest(t, f.server, "PUT", path+"/draw", confirm, f.organizer.AccessToken, 200)
	return bracket
}
func (f competitionFixture) bracket(t *testing.T) contract.Bracket {
	t.Helper()
	return integrationDecode[contract.Bracket](t, integrationRequest(t, f.server, "GET", BaseURL+"/tournaments/"+f.tournament.String()+"/bracket", "", "", 200))
}
func (f competitionFixture) start(t *testing.T, m contract.Match) contract.Match {
	t.Helper()
	if _, err := f.db.Exec(context.Background(), `UPDATE matches SET starts_at=now()-interval '1 hour' WHERE id=$1`, m.Id); err != nil {
		t.Fatal(err)
	}
	return m
}
func resultBody(t *testing.T, m contract.Match, home, away int) string {
	t.Helper()
	return stringMustJSON(t, contract.ReportResultRequest{Version: m.Version, HomeScore: home, AwayScore: away, Games: []contract.GameScore{{Number: 1, HomeScore: home, AwayScore: away}}})
}
func (f competitionFixture) refereeResult(t *testing.T, m contract.Match, home, away int) contract.Match {
	t.Helper()
	f.start(t, m)
	return integrationDecode[contract.Match](t, integrationRequest(t, f.server, "PUT", BaseURL+"/matches/"+m.Id.String()+"/referee-result", resultBody(t, m, home, away), f.referee.AccessToken, 200))
}
func TestCompetitionFormatsIntegration(t *testing.T) {
	for _, tc := range []struct {
		format contract.TournamentFormat
		n      int
		total  int
		byes   int
	}{
		{contract.TournamentFormatSingleElimination, 3, 3, 1},
		{contract.TournamentFormatRoundRobin, 3, 3, 0},
		{contract.TournamentFormatGroupsPlayoffs, 5, 5, 0},
	} {
		t.Run(string(tc.format), func(t *testing.T) {
			f := newCompetitionFixture(t, tc.format, tc.n)
			b := f.draw(t)
			byes := 0
			for _, m := range b.Matches {
				if m.Status == contract.MatchStatusBye {
					byes++
				}
			}
			if byes != tc.byes {
				t.Fatalf("byes=%d", byes)
			}
			integrationRequest(t, f.server, "GET", BaseURL+"/tournaments/"+f.tournament.String()+"/standings", "", "", 200)
			for iteration := 0; iteration < tc.total+1; iteration++ {
				found := false
				for _, m := range b.Matches {
					if m.Status == contract.MatchStatusScheduled && m.HomeApplicationId != nil && m.AwayApplicationId != nil {
						f.refereeResult(t, m, 1, 0)
						found = true
						break
					}
				}
				if !found {
					break
				}
				b = f.bracket(t)
			}
			completed, all := 0, len(b.Matches)
			for _, m := range b.Matches {
				if m.Status == contract.MatchStatusCompleted || m.Status == contract.MatchStatusBye {
					completed++
				}
			}
			if completed != all || all != tc.total {
				t.Fatalf("completed=%d all=%d want=%d", completed, all, tc.total)
			}
			tour := integrationDecode[contract.Tournament](t, integrationRequest(t, f.server, "GET", BaseURL+"/tournaments/"+f.tournament.String(), "", "", 200))
			if tour.Status != contract.TournamentStatusCompleted {
				t.Fatal(tour.Status)
			}
			var sum float64
			var played, placed int
			if err := f.db.QueryRow(context.Background(), `SELECT coalesce(sum(value),0),coalesce(sum(played),0) FROM ratings WHERE discipline_id=$1`, f.discipline).Scan(&sum, &played); err != nil {
				t.Fatal(err)
			}
			if math.Abs(sum-float64(tc.n)*1500) > 1e-7 || played != (tc.total-tc.byes)*2 {
				t.Fatalf("ratings sum=%f played=%d", sum, played)
			}
			if err := f.db.QueryRow(context.Background(), `SELECT count(*) FROM applications WHERE tournament_id=$1 AND final_place IS NOT NULL`, f.tournament).Scan(&placed); err != nil || placed != tc.n {
				t.Fatalf("placed=%d err=%v", placed, err)
			}
		})
	}
}
func TestResultConfirmationDisputeAndCorrectionIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatRoundRobin, 3)
	b := f.draw(t)
	m := f.start(t, b.Matches[0])
	base := BaseURL + "/matches/" + m.Id.String()
	captain, opponent := f.captains[*m.HomeApplicationId], f.captains[*m.AwayApplicationId]
	r := integrationDecode[contract.ResultReport](t, integrationRequest(t, f.server, "POST", base+"/reports", resultBody(t, m, 2, 1), captain.AccessToken, 201))
	confirmPath := base + "/reports/" + r.Id.String() + "/confirm"
	integrationRequest(t, f.server, "POST", confirmPath, `{"version":2}`, captain.AccessToken, 403)
	integrationRequest(t, f.server, "POST", confirmPath, `{"version":1}`, opponent.AccessToken, 409)
	m = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "POST", confirmPath, `{"version":2}`, opponent.AccessToken, 200))
	integrationRequest(t, f.server, "POST", confirmPath, `{"version":2}`, opponent.AccessToken, 200)
	// Следующий результат уже зависит от рейтинга первого, поэтому исправление
	// должно изменить обе записи истории, а не прибавить компенсационную дельту.
	b = f.bracket(t)
	for _, next := range b.Matches {
		if next.Id != m.Id {
			f.refereeResult(t, next, 1, 0)
			break
		}
	}
	integrationRequest(t, f.server, "POST", base+"/reports/"+r.Id.String()+"/dispute", stringMustJSON(t, contract.DisputeRequest{Version: m.Version, Reason: "Протокол содержит обратный счёт"}), opponent.AccessToken, 201)
	m = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", base, "", "", 200))
	resolution := contract.ResolveDisputeRequest{Version: m.Version, HomeScore: 1, AwayScore: 2, Games: []contract.GameScore{{Number: 1, HomeScore: 1, AwayScore: 2}}, Reason: "Проверен исходный протокол"}
	integrationRequest(t, f.server, "POST", base+"/dispute/resolve", stringMustJSON(t, resolution), captain.AccessToken, 403)
	integrationRequest(t, f.server, "POST", base+"/dispute/resolve", stringMustJSON(t, resolution), f.referee.AccessToken, 200)
	var revisions int
	if err := f.db.QueryRow(context.Background(), `SELECT count(*) FROM rating_changes WHERE revision>1`).Scan(&revisions); err != nil || revisions < 2 {
		t.Fatalf("revisions=%d err=%v", revisions, err)
	}
	integrationRequest(t, f.server, "GET", base+"/reports", "", captain.AccessToken, 200)
	integrationRequest(t, f.server, "GET", base+"/dispute", "", opponent.AccessToken, 200)
	integrationRequest(t, f.server, "GET", BaseURL+"/disputes?status=resolved", "", f.referee.AccessToken, 200)
}

func TestCancelledTournamentCannotResolveDisputeIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatSingleElimination, 2)
	bracket := f.draw(t)
	m := f.start(t, bracket.Matches[0])
	base := BaseURL + "/matches/" + m.Id.String()
	captain, opponent := f.captains[*m.HomeApplicationId], f.captains[*m.AwayApplicationId]
	report := integrationDecode[contract.ResultReport](t, integrationRequest(t, f.server, "POST", base+"/reports", resultBody(t, m, 1, 0), captain.AccessToken, 201))
	dispute := integrationDecode[contract.Dispute](t, integrationRequest(t, f.server, "POST", base+"/reports/"+report.Id.String()+"/dispute", stringMustJSON(t, contract.DisputeRequest{
		Version: m.Version + 1,
		Reason:  "Счёт оспаривается",
	}), opponent.AccessToken, 201))
	m = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", base, "", "", 200))
	tournamentPath := BaseURL + "/tournaments/" + f.tournament.String()
	tournament := integrationDecode[contract.Tournament](t, integrationRequest(t, f.server, "GET", tournamentPath, "", "", 200))
	integrationRequest(t, f.server, "POST", tournamentPath+"/cancel", stringMustJSON(t, contract.ReasonRequest{
		Version: tournament.Version,
		Reason:  "Турнир отменён организатором",
	}), f.organizer.AccessToken, 200)

	resolution := contract.ResolveDisputeRequest{
		Version:   m.Version,
		HomeScore: 1,
		AwayScore: 0,
		Games:     []contract.GameScore{{Number: 1, HomeScore: 1, AwayScore: 0}},
		Reason:    "Попытка разрешить спор после отмены",
	}
	integrationRequest(t, f.server, "POST", base+"/dispute/resolve", stringMustJSON(t, resolution), f.referee.AccessToken, 409)

	tournament = integrationDecode[contract.Tournament](t, integrationRequest(t, f.server, "GET", tournamentPath, "", "", 200))
	if tournament.Status != contract.TournamentStatusCancelled {
		t.Fatalf("tournament status=%s, want=cancelled", tournament.Status)
	}
	unchanged := integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", base, "", "", 200))
	if unchanged.Status != contract.MatchStatusDisputed || unchanged.Version != m.Version || unchanged.HomeScore != nil || unchanged.AwayScore != nil {
		t.Fatalf("cancelled tournament match changed: %+v", unchanged)
	}
	open := integrationDecode[contract.Dispute](t, integrationRequest(t, f.server, "GET", base+"/dispute", "", opponent.AccessToken, 200))
	if open.Id != dispute.Id || open.Status != contract.DisputeStatusOpen || open.Resolution != nil || open.ResolvedBy != nil {
		t.Fatalf("dispute changed after rejected resolution: %+v", open)
	}
	reports := integrationDecode[[]contract.ResultReport](t, integrationRequest(t, f.server, "GET", base+"/reports", "", captain.AccessToken, 200))
	if len(reports) != 1 || reports[0].Id != report.Id || reports[0].Status != contract.ResultReportStatusDisputed {
		t.Fatalf("reports changed after rejected resolution: %+v", reports)
	}
	var ratings, revisions int
	if err := f.db.QueryRow(context.Background(), `SELECT (SELECT count(*) FROM ratings WHERE discipline_id=$1),(SELECT count(*) FROM rating_changes WHERE tournament_id=$2)`, f.discipline, f.tournament).Scan(&ratings, &revisions); err != nil {
		t.Fatal(err)
	}
	if ratings != 0 || revisions != 0 {
		t.Fatalf("cancelled tournament affected ratings: ratings=%d revisions=%d", ratings, revisions)
	}
}

func TestCompletedTournamentCanResolveDisputeIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatSingleElimination, 2)
	bracket := f.draw(t)
	m := f.refereeResult(t, bracket.Matches[0], 1, 0)
	base := BaseURL + "/matches/" + m.Id.String()
	tournamentPath := BaseURL + "/tournaments/" + f.tournament.String()
	tournament := integrationDecode[contract.Tournament](t, integrationRequest(t, f.server, "GET", tournamentPath, "", "", 200))
	if tournament.Status != contract.TournamentStatusCompleted {
		t.Fatalf("tournament status=%s, want=completed", tournament.Status)
	}
	opponent := f.captains[*m.AwayApplicationId]
	reports := integrationDecode[[]contract.ResultReport](t, integrationRequest(t, f.server, "GET", base+"/reports", "", opponent.AccessToken, 200))
	integrationRequest(t, f.server, "POST", base+"/reports/"+reports[0].Id.String()+"/dispute", stringMustJSON(t, contract.DisputeRequest{
		Version: m.Version,
		Reason:  "Проверка итогового протокола",
	}), opponent.AccessToken, 201)
	m = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", base, "", "", 200))
	resolution := contract.ResolveDisputeRequest{
		Version:   m.Version,
		HomeScore: 0,
		AwayScore: 1,
		Games:     []contract.GameScore{{Number: 1, HomeScore: 0, AwayScore: 1}},
		Reason:    "Судья исправил завершённый результат",
	}
	corrected := integrationDecode[contract.Match](t, integrationRequest(t, f.server, "POST", base+"/dispute/resolve", stringMustJSON(t, resolution), f.referee.AccessToken, 200))
	if corrected.Status != contract.MatchStatusCompleted || corrected.WinnerApplicationId == nil || *corrected.WinnerApplicationId != *m.AwayApplicationId {
		t.Fatalf("completed tournament result was not corrected: %+v", corrected)
	}
}

func TestPendingDisputeAutoConfirmationAndScheduleIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatRoundRobin, 3)
	b := f.draw(t)
	m, next := b.Matches[0], b.Matches[1]
	base := BaseURL + "/matches/" + m.Id.String()
	start := time.Now().UTC().Add(-2 * time.Hour)
	venue := "Hall"
	schedule := contract.ScheduleMatchRequest{Version: m.Version, StartsAt: start, Venue: &venue}
	m = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "PUT", base+"/schedule", stringMustJSON(t, schedule), f.organizer.AccessToken, 200))
	integrationRequest(t, f.server, "PUT", base+"/schedule", stringMustJSON(t, schedule), f.organizer.AccessToken, 409)
	schedule.Version = next.Version
	integrationRequest(t, f.server, "PUT", BaseURL+"/matches/"+next.Id.String()+"/schedule", stringMustJSON(t, schedule), f.organizer.AccessToken, 409)
	captain, opponent := f.captains[*m.HomeApplicationId], f.captains[*m.AwayApplicationId]
	r := integrationDecode[contract.ResultReport](t, integrationRequest(t, f.server, "POST", base+"/reports", resultBody(t, m, 1, 0), captain.AccessToken, 201))
	integrationRequest(t, f.server, "POST", base+"/reports/"+r.Id.String()+"/dispute", stringMustJSON(t, contract.DisputeRequest{Version: m.Version + 1, Reason: "Счёт оспаривается"}), opponent.AccessToken, 201)
	if err := service.New(postgres.New(f.db), config.Config{}, nil).AutoConfirm(context.Background()); err != nil {
		t.Fatal(err)
	}
	disputed := integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", base, "", "", 200))
	if disputed.Status != contract.MatchStatusDisputed {
		t.Fatal(disputed.Status)
	}
	req := contract.ResolveDisputeRequest{Version: disputed.Version, HomeScore: 1, AwayScore: 0, Games: []contract.GameScore{{Number: 1, HomeScore: 1, AwayScore: 0}}, Reason: "Результат подтверждён судьёй"}
	integrationRequest(t, f.server, "POST", base+"/dispute/resolve", stringMustJSON(t, req), f.referee.AccessToken, 200)
	next = f.start(t, next)
	base = BaseURL + "/matches/" + next.Id.String()
	captain = f.captains[*next.HomeApplicationId]
	r = integrationDecode[contract.ResultReport](t, integrationRequest(t, f.server, "POST", base+"/reports", resultBody(t, next, 1, 0), captain.AccessToken, 201))
	if _, err := f.db.Exec(context.Background(), `UPDATE result_reports SET confirmation_deadline=now()-interval '1 second' WHERE id=$1`, r.Id); err != nil {
		t.Fatal(err)
	}
	if err := service.New(postgres.New(f.db), config.Config{}, nil).AutoConfirm(context.Background()); err != nil {
		t.Fatal(err)
	}
	confirmed := integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", base, "", "", 200))
	if confirmed.Status != contract.MatchStatusCompleted {
		t.Fatal(confirmed.Status)
	}
}

func TestPlayoffCorrectionCannotChangeStartedDependentIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatSingleElimination, 4)
	b := f.draw(t)
	m := f.refereeResult(t, b.Matches[0], 1, 0)
	base := BaseURL + "/matches/" + m.Id.String()
	if _, err := f.db.Exec(context.Background(), `UPDATE matches SET starts_at=now()-interval '1 minute' WHERE id=$1`, m.NextMatchId); err != nil {
		t.Fatal(err)
	}
	reports := integrationDecode[[]contract.ResultReport](t, integrationRequest(t, f.server, "GET", base+"/reports", "", f.captains[*m.HomeApplicationId].AccessToken, 200))
	integrationRequest(t, f.server, "POST", base+"/reports/"+reports[0].Id.String()+"/dispute", stringMustJSON(t, contract.DisputeRequest{Version: m.Version, Reason: "Проверка счёта"}), f.captains[*m.AwayApplicationId].AccessToken, 201)
	m = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", base, "", "", 200))
	req := contract.ResolveDisputeRequest{Version: m.Version, HomeScore: 0, AwayScore: 1, Games: []contract.GameScore{{Number: 1, HomeScore: 0, AwayScore: 1}}, Reason: "Изменение победителя"}
	integrationRequest(t, f.server, "POST", base+"/dispute/resolve", stringMustJSON(t, req), f.referee.AccessToken, 409)
	// Коррекция счёта с прежним победителем не меняет состав следующего матча.
	req.HomeScore = 2
	req.AwayScore = 0
	req.Games = []contract.GameScore{{Number: 1, HomeScore: 2, AwayScore: 0}}
	integrationRequest(t, f.server, "POST", base+"/dispute/resolve", stringMustJSON(t, req), f.referee.AccessToken, 200)
}
func TestScheduleRoundAndMatchRefereeIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatRoundRobin, 4)
	b := f.draw(t)
	req := contract.ScheduleRoundRequest{Version: 2, StageId: b.Stages[0].Id, Round: 1, StartsAt: time.Now().UTC().Add(2 * time.Hour), IntervalMinutes: 60, Venue: "Main hall"}
	matches := integrationDecode[[]contract.Match](t, integrationRequest(t, f.server, "POST", BaseURL+"/tournaments/"+f.tournament.String()+"/schedule-round", stringMustJSON(t, req), f.organizer.AccessToken, 200))
	if len(matches) != 2 || matches[0].Version != 2 || matches[1].Version != 2 {
		t.Fatal(matches)
	}
	if matches[1].StartsAt.Sub(*matches[0].StartsAt) != time.Hour {
		t.Fatal("wrong round interval")
	}
	judgeID, judge := seedIntegrationUser(t, f.db, "second_referee", contract.Referee)
	assign := contract.AssignRefereeRequest{Version: matches[0].Version, RefereeId: judgeID}
	path := BaseURL + "/matches/" + matches[0].Id.String() + "/referee"
	integrationRequest(t, f.server, "PUT", path, stringMustJSON(t, assign), judge.AccessToken, 403)
	assigned := integrationDecode[contract.Match](t, integrationRequest(t, f.server, "PUT", path, stringMustJSON(t, assign), f.organizer.AccessToken, 200))
	if assigned.RefereeId == nil || *assigned.RefereeId != judgeID {
		t.Fatal(assigned)
	}
	integrationRequest(t, f.server, "PUT", path, stringMustJSON(t, assign), f.organizer.AccessToken, 409)
}
func TestDrawRejectsChangedAndInvalidParticipantsIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatRoundRobin, 3)
	base := BaseURL + "/tournaments/" + f.tournament.String()
	duplicate := []uuid.UUID{f.applications[0], uuid.New(), f.applications[2]}
	req := contract.SeedingRequest{Method: contract.SeedingRequestMethodManual, Version: 1, ApplicationIds: &duplicate}
	integrationRequest(t, f.server, "POST", base+"/draw-previews", stringMustJSON(t, req), f.organizer.AccessToken, 422)
	req.ApplicationIds = &f.applications
	preview := integrationDecode[contract.DrawPreview](t, integrationRequest(t, f.server, "POST", base+"/draw-previews", stringMustJSON(t, req), f.organizer.AccessToken, 201))
	if _, err := f.db.Exec(context.Background(), `UPDATE applications SET status='withdrawn' WHERE id=$1`, f.applications[0]); err != nil {
		t.Fatal(err)
	}
	integrationRequest(t, f.server, "PUT", base+"/draw", stringMustJSON(t, contract.ConfirmDrawRequest{PreviewId: preview.Id, Version: 1}), f.organizer.AccessToken, 409)
}

func TestConfirmationDisputeRaceIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatRoundRobin, 2)
	b := f.draw(t)
	m := f.start(t, b.Matches[0])
	base := BaseURL + "/matches/" + m.Id.String()
	captain, opponent := f.captains[*m.HomeApplicationId], f.captains[*m.AwayApplicationId]
	report := integrationDecode[contract.ResultReport](t, integrationRequest(t, f.server, "POST", base+"/reports", resultBody(t, m, 1, 0), captain.AccessToken, 201))
	type response struct {
		path    string
		rec     *httptest.ResponseRecorder
		request *http.Request
	}
	results := make(chan response, 2)
	start := make(chan struct{})
	for _, action := range []string{"confirm", "dispute"} {
		go func(action string) {
			body := `{"version":2}`
			if action == "dispute" {
				body = `{"version":2,"reason":"Проверка гонки подтверждения"}`
			}
			path := base + "/reports/" + report.Id.String() + "/" + action
			req := httptest.NewRequest("POST", path, strings.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+opponent.AccessToken)
			rec := httptest.NewRecorder()
			<-start
			f.server.ServeHTTP(rec, req)
			results <- response{path, rec, req}
		}(action)
	}
	close(start)
	success, conflicts := 0, 0
	for range 2 {
		r := <-results
		validateIntegrationResponse(t, r.request, r.rec)
		if r.rec.Code == 200 || r.rec.Code == 201 {
			success++
		} else if r.rec.Code == 409 {
			conflicts++
		} else {
			t.Fatalf("unexpected race response %s: %d %s", r.path, r.rec.Code, r.rec.Body.String())
		}
	}
	if success != 1 || conflicts != 1 {
		t.Fatalf("success=%d conflict=%d", success, conflicts)
	}
}

func TestCaptainReplacementKeepsResultSideAndReadAccessIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatRoundRobin, 2)
	b := f.draw(t)
	m := f.start(t, b.Matches[0])
	base := BaseURL + "/matches/" + m.Id.String()
	captain, opponent := f.captains[*m.HomeApplicationId], f.captains[*m.AwayApplicationId]
	r := integrationDecode[contract.ResultReport](t, integrationRequest(t, f.server, "POST", base+"/reports", resultBody(t, m, 1, 0), captain.AccessToken, 201))
	replacementID, replacement := seedIntegrationUser(t, f.db, "new_captain")
	ctx := context.Background()
	tx, err := f.db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	var teamID, oldCaptain uuid.UUID
	if err = tx.QueryRow(ctx, `SELECT t.id,t.captain_id FROM teams t JOIN applications a ON a.team_id=t.id WHERE a.id=$1`, m.HomeApplicationId).Scan(&teamID, &oldCaptain); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, teamID, replacementID); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `UPDATE teams SET captain_id=$2 WHERE id=$1`, teamID, replacementID); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, teamID, oldCaptain); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	integrationRequest(t, f.server, "GET", base+"/reports", "", replacement.AccessToken, 200)
	path := base + "/reports/" + r.Id.String() + "/confirm"
	integrationRequest(t, f.server, "POST", path, `{"version":2}`, replacement.AccessToken, 403)
	integrationRequest(t, f.server, "POST", path, `{"version":2}`, opponent.AccessToken, 200)
	var notifications int
	if err = f.db.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE user_id=$1 AND resource_id=$2 AND kind='result_confirmed'`, replacementID, m.Id).Scan(&notifications); err != nil || notifications != 1 {
		t.Fatalf("replacement captain notifications=%d err=%v", notifications, err)
	}
}

func cloneExternalMatch(t *testing.T, f competitionFixture, home, away uuid.UUID, start time.Time) (uuid.UUID, uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	tid, sid, mid := uuid.New(), uuid.New(), uuid.New()
	ha, aa := uuid.New(), uuid.New()
	tx, err := f.db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `INSERT INTO tournaments(id,organizer_id,discipline_id,name,format,status,max_participants,registration_opens_at,registration_closes_at,starts_at,ends_at,timezone,best_of,round_robin_legs,win_points,draw_points,loss_points,tie_breakers,min_roster_size,max_roster_size,allows_draw,score_unit) SELECT $2,organizer_id,discipline_id,'External','round_robin','in_progress',2,registration_opens_at,registration_closes_at,starts_at,ends_at,timezone,1,1,win_points,draw_points,loss_points,tie_breakers,min_roster_size,max_roster_size,allows_draw,score_unit FROM tournaments WHERE id=$1`, f.tournament, tid)
	if err != nil {
		t.Fatal(err)
	}
	for _, pair := range [][2]uuid.UUID{{home, ha}, {away, aa}} {
		if _, err = tx.Exec(ctx, `INSERT INTO applications(id,tournament_id,team_id,submitted_by,status) SELECT $2,$3,team_id,submitted_by,'accepted' FROM applications WHERE id=$1`, pair[0], pair[1], tid); err != nil {
			t.Fatal(err)
		}
		if _, err = tx.Exec(ctx, `INSERT INTO application_players(application_id,tournament_id,user_id) SELECT $2,$3,user_id FROM application_players WHERE application_id=$1`, pair[0], pair[1], tid); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO stages(id,tournament_id,kind,position,status) VALUES($1,$2,'round_robin',1,'in_progress')`, sid, tid); err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO matches(id,tournament_id,stage_id,round,position,home_application_id,away_application_id,starts_at,venue) VALUES($1,$2,$3,1,1,$4,$5,$6,'External hall')`, mid, tid, sid, ha, aa, start); err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	return tid, mid
}
func TestPlayoffScheduleDependenciesAndAutomaticUnscheduleIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatSingleElimination, 4)
	b := f.draw(t)
	semi := b.Matches[0]
	final := b.Matches[len(b.Matches)-1]
	start := time.Now().UTC().Add(2 * time.Hour)
	venue := "Arena"
	schedule := contract.ScheduleMatchRequest{Version: semi.Version, StartsAt: start, Venue: &venue}
	semi = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "PUT", BaseURL+"/matches/"+semi.Id.String()+"/schedule", stringMustJSON(t, schedule), f.organizer.AccessToken, 200))
	schedule.Version = final.Version
	integrationRequest(t, f.server, "PUT", BaseURL+"/matches/"+final.Id.String()+"/schedule", stringMustJSON(t, schedule), f.organizer.AccessToken, 409)
	schedule.StartsAt = start.Add(time.Hour)
	final = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "PUT", BaseURL+"/matches/"+final.Id.String()+"/schedule", stringMustJSON(t, schedule), f.organizer.AccessToken, 200))
	cloneExternalMatch(t, f, *semi.HomeApplicationId, *semi.AwayApplicationId, schedule.StartsAt)
	f.refereeResult(t, semi, 1, 0)
	final = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", BaseURL+"/matches/"+final.Id.String(), "", "", 200))
	if final.HomeApplicationId == nil || final.StartsAt != nil {
		t.Fatalf("advanced final has invalid schedule: %+v", final)
	}
	var messages int
	if err := f.db.QueryRow(context.Background(), `SELECT count(*) FROM notifications WHERE resource_id=$1 AND kind='match_unscheduled'`, final.Id).Scan(&messages); err != nil || messages == 0 {
		t.Fatalf("unschedule messages=%d err=%v", messages, err)
	}
}
func TestAutoConfirmationSkipsCancelledTournamentIntegration(t *testing.T) {
	f := newCompetitionFixture(t, contract.TournamentFormatRoundRobin, 2)
	b := f.draw(t)
	m := f.start(t, b.Matches[0])
	ctx := context.Background()
	tid, external := cloneExternalMatch(t, f, *m.HomeApplicationId, *m.AwayApplicationId, time.Now().UTC().Add(-time.Hour))
	captain := f.captains[*m.HomeApplicationId]
	cancelledReport := integrationDecode[contract.ResultReport](t, integrationRequest(t, f.server, "POST", BaseURL+"/matches/"+external.String()+"/reports", resultBody(t, contract.Match{Version: 1}, 1, 0), captain.AccessToken, 201))
	activeReport := integrationDecode[contract.ResultReport](t, integrationRequest(t, f.server, "POST", BaseURL+"/matches/"+m.Id.String()+"/reports", resultBody(t, m, 1, 0), captain.AccessToken, 201))
	if _, err := f.db.Exec(ctx, `UPDATE result_reports SET confirmation_deadline=now()-interval '2 hours' WHERE id=$1`, cancelledReport.Id); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(ctx, `UPDATE result_reports SET confirmation_deadline=now()-interval '1 hour' WHERE id=$1`, activeReport.Id); err != nil {
		t.Fatal(err)
	}
	if _, err := f.db.Exec(ctx, `UPDATE tournaments SET status='cancelled' WHERE id=$1`, tid); err != nil {
		t.Fatal(err)
	}
	if err := service.New(postgres.New(f.db), config.Config{}, nil).AutoConfirm(ctx); err != nil {
		t.Fatal(err)
	}
	active := integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", BaseURL+"/matches/"+m.Id.String(), "", "", 200))
	if active.Status != contract.MatchStatusCompleted {
		t.Fatal(active.Status)
	}
	var status string
	if err := f.db.QueryRow(ctx, `SELECT status FROM result_reports WHERE id=$1`, cancelledReport.Id).Scan(&status); err != nil || status != "pending" {
		t.Fatalf("cancelled report=%s err=%v", status, err)
	}
}

func TestGroupCorrectionReordersSamePlayoffQualifiersIntegration(t *testing.T) {
	for _, started := range []bool{false, true} {
		t.Run(fmt.Sprintf("playoffs_started_%t", started), func(t *testing.T) {
			f := newCompetitionFixture(t, contract.TournamentFormatGroupsPlayoffs, 6)
			if _, err := f.db.Exec(context.Background(), `UPDATE tournaments SET qualifiers_per_group=2 WHERE id=$1`, f.tournament); err != nil {
				t.Fatal(err)
			}
			b := f.draw(t)
			groupStage := b.Stages[0].Id
			group := b.Groups[0].Id
			var corrected contract.Match
			for _, m := range b.Matches {
				done := f.refereeResult(t, m, 1, 0)
				if m.GroupId != nil && *m.GroupId == group && m.Round == 2 {
					corrected = done
				}
			}
			b = f.bracket(t)
			old := map[uuid.UUID]bool{}
			var playoff contract.Match
			for _, m := range b.Matches {
				if m.StageId != groupStage {
					old[m.Id] = true
					playoff = m
				}
			}
			if len(old) != 3 {
				t.Fatal("four qualifiers should produce three playoff matches")
			}
			if started {
				if _, err := f.db.Exec(context.Background(), `UPDATE matches SET starts_at=now()-interval '1 minute' WHERE id=$1`, playoff.Id); err != nil {
					t.Fatal(err)
				}
			}
			base := BaseURL + "/matches/" + corrected.Id.String()
			reports := integrationDecode[[]contract.ResultReport](t, integrationRequest(t, f.server, "GET", base+"/reports", "", f.captains[*corrected.AwayApplicationId].AccessToken, 200))
			integrationRequest(t, f.server, "POST", base+"/reports/"+reports[0].Id.String()+"/dispute", stringMustJSON(t, contract.DisputeRequest{Version: corrected.Version, Reason: "Пересмотр результата лидеров группы"}), f.captains[*corrected.AwayApplicationId].AccessToken, 201)
			corrected = integrationDecode[contract.Match](t, integrationRequest(t, f.server, "GET", base, "", "", 200))
			resolution := contract.ResolveDisputeRequest{Version: corrected.Version, HomeScore: 0, AwayScore: 1, Games: []contract.GameScore{{Number: 1, HomeScore: 0, AwayScore: 1}}, Reason: "Первое и второе места в группе меняются"}
			want := 200
			if started {
				want = 409
			}
			integrationRequest(t, f.server, "POST", base+"/dispute/resolve", stringMustJSON(t, resolution), f.referee.AccessToken, want)
			after := f.bracket(t)
			same := 0
			for _, m := range after.Matches {
				if old[m.Id] {
					same++
				}
			}
			if started && same != 3 {
				t.Fatal("started playoffs were rebuilt")
			}
			if !started && same != 0 {
				t.Fatal("changed seed order kept old playoff pairings")
			}
		})
	}
}
