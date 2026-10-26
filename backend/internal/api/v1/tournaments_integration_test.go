//go:build integration

package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func validTournamentRequest() contract.CreateTournamentRequest {
	return contract.CreateTournamentRequest{
		Name: "University Cup", Format: contract.TournamentFormatSingleElimination,
		MaxParticipants: 8, Timezone: "Europe/Moscow",
		Rules: contract.Rules{BestOf: 1, RoundRobinLegs: 1, WinPoints: 3, DrawPoints: 1,
			TieBreakers: []contract.TieBreaker{contract.TieBreakerScoreDifference}, ConfirmationTimeoutHours: 24},
	}
}

func tournamentBody(t *testing.T, body any) string {
	t.Helper()
	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

func tournamentResponse[T any](t *testing.T, body []byte) T {
	t.Helper()
	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("invalid response: %v: %s", err, body)
	}
	return result
}

func seedTournamentDiscipline(t *testing.T, db *pgxpool.Pool) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_, err := db.Exec(context.Background(), `INSERT INTO disciplines(id,name,min_roster_size,max_roster_size,allows_draw,score_unit) VALUES($1,$2,1,1,false,'points')`, id, "Chess "+id.String())
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func createIntegrationTournament(t *testing.T, server *echo.Echo, discipline uuid.UUID, token string, limit int) contract.Tournament {
	t.Helper()
	r := validTournamentRequest()
	r.DisciplineId, r.MaxParticipants = discipline, limit
	now := time.Now().UTC()
	opens, closes, ends := now.Add(-time.Hour), now.Add(time.Hour), now.Add(48*time.Hour)
	r.RegistrationOpensAt, r.RegistrationClosesAt, r.StartsAt, r.EndsAt = &opens, &closes, &closes, &ends
	response := integrationRequest(t, server, "POST", "/api/v1/tournaments", tournamentBody(t, r), token, 201)
	return tournamentResponse[contract.Tournament](t, response.Body.Bytes())
}

func publishIntegrationTournament(t *testing.T, server *echo.Echo, tournament contract.Tournament, token string) contract.Tournament {
	t.Helper()
	response := integrationRequest(t, server, "POST", "/api/v1/tournaments/"+tournament.Id.String()+"/publish", tournamentBody(t, contract.VersionRequest{Version: tournament.Version}), token, 200)
	return tournamentResponse[contract.Tournament](t, response.Body.Bytes())
}

func createApplicationTeam(t *testing.T, server *echo.Echo, name, token string) contract.Team {
	t.Helper()
	response := integrationRequest(t, server, "POST", "/api/v1/teams", tournamentBody(t, contract.CreateTeamRequest{Name: name}), token, 201)
	return tournamentResponse[contract.Team](t, response.Body.Bytes())
}

func submitIntegrationApplication(t *testing.T, server *echo.Echo, tournamentID uuid.UUID, teamID *uuid.UUID, playerID uuid.UUID, token string) contract.Application {
	t.Helper()
	response := integrationRequest(t, server, "POST", "/api/v1/tournaments/"+tournamentID.String()+"/applications", tournamentBody(t, contract.ApplicationRequest{TeamId: teamID, PlayerIds: []uuid.UUID{playerID}}), token, 201)
	return tournamentResponse[contract.Application](t, response.Body.Bytes())
}

func reviewIntegrationApplication(t *testing.T, server *echo.Echo, application contract.Application, token string, request contract.ReviewApplicationRequest, want int) contract.Application {
	t.Helper()
	path := "/api/v1/tournaments/" + application.TournamentId.String() + "/applications/" + application.Id.String() + "/review"
	response := integrationRequest(t, server, "POST", path, tournamentBody(t, request), token, want)
	if want != 200 {
		return contract.Application{}
	}
	return tournamentResponse[contract.Application](t, response.Body.Bytes())
}

func TestTournamentApplicationsIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	_, organizer := seedIntegrationUser(t, db, "organizer", contract.Organizer)
	playerA, tokensA := seedIntegrationUser(t, db, "playerA", contract.Participant)
	playerB, tokensB := seedIntegrationUser(t, db, "playerB", contract.Participant)
	playerC, tokensC := seedIntegrationUser(t, db, "playerC", contract.Participant)
	discipline := seedTournamentDiscipline(t, db)
	request := validTournamentRequest()
	request.DisciplineId = discipline
	integrationRequest(t, server, "POST", "/api/v1/tournaments", tournamentBody(t, request), tokensA.AccessToken, 403)
	tournament := createIntegrationTournament(t, server, discipline, organizer.AccessToken, 2)
	path := "/api/v1/tournaments/" + tournament.Id.String()
	integrationRequest(t, server, "GET", path, "", "", 404)
	mine := integrationRequest(t, server, "GET", "/api/v1/users/me/tournaments", "", organizer.AccessToken, 200)
	if page := tournamentResponse[contract.TournamentPage](t, mine.Body.Bytes()); len(page.Items) != 1 || page.Items[0].Status != contract.TournamentStatusDraft {
		t.Fatalf("organizer lost draft: %+v", page)
	}
	integrationRequest(t, server, "POST", path+"/publish", `{"version":999}`, organizer.AccessToken, 409)
	tournament = publishIntegrationTournament(t, server, tournament, organizer.AccessToken)
	integrationRequest(t, server, "PATCH", path, `{"version":1,"name":"Stale"}`, organizer.AccessToken, 409)
	integrationRequest(t, server, "PATCH", path, tournamentBody(t, map[string]any{"version": tournament.Version, "name": "Foreign edit"}), tokensA.AccessToken, 403)
	integrationRequest(t, server, "GET", path, "", "", 200)
	teamA := createApplicationTeam(t, server, "A", tokensA.AccessToken)
	teamB := createApplicationTeam(t, server, "B", tokensB.AccessToken)
	teamC := createApplicationTeam(t, server, "C", tokensC.AccessToken)
	teamAOther := createApplicationTeam(t, server, "A other", tokensA.AccessToken)
	applicationA := submitIntegrationApplication(t, server, tournament.Id, &teamA.Id, playerA, tokensA.AccessToken)
	integrationRequest(t, server, "POST", path+"/applications", tournamentBody(t, contract.ApplicationRequest{TeamId: &teamAOther.Id, PlayerIds: []uuid.UUID{playerA}}), tokensA.AccessToken, 409)
	integrationRequest(t, server, "POST", path+"/applications", tournamentBody(t, contract.ApplicationRequest{TeamId: &teamA.Id, PlayerIds: []uuid.UUID{playerB}}), tokensA.AccessToken, 422)
	accept := contract.ReviewApplicationRequest{Decision: contract.ReviewApplicationRequestDecisionAccept, Version: applicationA.Version}
	reviewIntegrationApplication(t, server, applicationA, tokensB.AccessToken, accept, 403)
	reviewIntegrationApplication(t, server, applicationA, organizer.AccessToken, contract.ReviewApplicationRequest{Decision: contract.ReviewApplicationRequestDecisionReject, Version: applicationA.Version}, 422)
	reason := "Incomplete registration details"
	applicationA = reviewIntegrationApplication(t, server, applicationA, organizer.AccessToken, contract.ReviewApplicationRequest{Decision: contract.ReviewApplicationRequestDecisionReject, Version: applicationA.Version, Reason: &reason}, 200)
	if applicationA.Status != contract.ApplicationStatusRejected {
		t.Fatalf("application not rejected: %+v", applicationA)
	}
	applicationA = submitIntegrationApplication(t, server, tournament.Id, &teamA.Id, playerA, tokensA.AccessToken)
	applicationB := submitIntegrationApplication(t, server, tournament.Id, &teamB.Id, playerB, tokensB.AccessToken)
	applicationC := submitIntegrationApplication(t, server, tournament.Id, &teamC.Id, playerC, tokensC.AccessToken)
	applicationA = reviewIntegrationApplication(t, server, applicationA, organizer.AccessToken, accept, 200)
	applicationB = reviewIntegrationApplication(t, server, applicationB, organizer.AccessToken, accept, 200)
	reviewIntegrationApplication(t, server, applicationC, organizer.AccessToken, accept, 409)
	integrationRequest(t, server, "POST", path+"/applications/"+applicationB.Id.String()+"/withdraw", tournamentBody(t, contract.VersionRequest{Version: applicationB.Version}), tokensB.AccessToken, 200)
	applicationC = reviewIntegrationApplication(t, server, applicationC, organizer.AccessToken, accept, 200)
	if applicationC.Status != contract.ApplicationStatusAccepted {
		t.Fatal("withdrawing did not release tournament capacity")
	}
	participants := integrationRequest(t, server, "GET", path+"/participants", "", "", 200)
	if page := tournamentResponse[contract.ApplicationPage](t, participants.Body.Bytes()); page.Total != 2 || len(page.Items) != 2 {
		t.Fatalf("wrong accepted participants: %+v", page)
	}
	integrationRequest(t, server, "GET", path+"/applications", "", tokensA.AccessToken, 403)
	integrationRequest(t, server, "GET", "/api/v1/users/me/applications", "", tokensA.AccessToken, 200)
	var claims int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM tournament_player_claims WHERE tournament_id=$1`, tournament.Id).Scan(&claims); err != nil || claims != 2 {
		t.Fatalf("claims after rejected/withdrawn applications: %d, %v", claims, err)
	}
}

func TestIndividualApplicationsAndDeadlinesIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	_, organizer := seedIntegrationUser(t, db, "individualOrganizer", contract.Organizer)
	player, tokens := seedIntegrationUser(t, db, "individualPlayer", contract.Participant)
	otherPlayer, otherTokens := seedIntegrationUser(t, db, "otherPlayer", contract.Participant)
	discipline := seedTournamentDiscipline(t, db)
	tournament := publishIntegrationTournament(t, server, createIntegrationTournament(t, server, discipline, organizer.AccessToken, 4), organizer.AccessToken)
	path := "/api/v1/tournaments/" + tournament.Id.String()
	application := submitIntegrationApplication(t, server, tournament.Id, nil, player, tokens.AccessToken)
	if !application.Team.Personal || len(application.Players) != 1 || application.Players[0].Id != player {
		t.Fatalf("individual team was not created: %+v", application)
	}
	personalTeam := application.Team.Id
	integrationRequest(t, server, "POST", path+"/applications/"+application.Id.String()+"/withdraw", tournamentBody(t, contract.VersionRequest{Version: application.Version}), tokens.AccessToken, 200)
	application = submitIntegrationApplication(t, server, tournament.Id, nil, player, tokens.AccessToken)
	if application.Team.Id != personalTeam {
		t.Fatal("individual reapplication created a second personal team")
	}
	_, err := db.Exec(context.Background(), `UPDATE tournaments SET registration_closes_at=now()-interval '1 second',starts_at=now() WHERE id=$1`, tournament.Id)
	if err != nil {
		t.Fatal(err)
	}
	integrationRequest(t, server, "POST", path+"/applications/"+application.Id.String()+"/withdraw", tournamentBody(t, contract.VersionRequest{Version: application.Version}), tokens.AccessToken, 409)
	integrationRequest(t, server, "POST", path+"/applications", tournamentBody(t, contract.ApplicationRequest{PlayerIds: []uuid.UUID{otherPlayer}}), otherTokens.AccessToken, 409)
	request := validTournamentRequest()
	request.DisciplineId = discipline
	draftResponse := integrationRequest(t, server, "POST", "/api/v1/tournaments", tournamentBody(t, request), organizer.AccessToken, 201)
	draft := tournamentResponse[contract.Tournament](t, draftResponse.Body.Bytes())
	integrationRequest(t, server, "POST", "/api/v1/tournaments/"+draft.Id.String()+"/publish", tournamentBody(t, contract.VersionRequest{Version: draft.Version}), organizer.AccessToken, 422)
}

func TestConcurrentApplicationReviewsRespectCapacityIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	_, organizer := seedIntegrationUser(t, db, "capacityOrganizer", contract.Organizer)
	discipline := seedTournamentDiscipline(t, db)
	tournament := publishIntegrationTournament(t, server, createIntegrationTournament(t, server, discipline, organizer.AccessToken, 2), organizer.AccessToken)
	applications := make([]contract.Application, 3)
	for index, nickname := range []string{"capacityFirst", "capacitySecond", "capacityThird"} {
		id, tokens := seedIntegrationUser(t, db, nickname, contract.Participant)
		applications[index] = submitIntegrationApplication(t, server, tournament.Id, nil, id, tokens.AccessToken)
	}
	accept := contract.ReviewApplicationRequest{Decision: contract.ReviewApplicationRequestDecisionAccept, Version: 1}
	reviewIntegrationApplication(t, server, applications[0], organizer.AccessToken, accept, 200)
	requests := make([]*http.Request, 2)
	responses := make([]*httptest.ResponseRecorder, 2)
	var workers sync.WaitGroup
	for index := range requests {
		application := applications[index+1]
		path := BaseURL + "/tournaments/" + tournament.Id.String() + "/applications/" + application.Id.String() + "/review"
		requests[index] = httptest.NewRequest("POST", path, strings.NewReader(tournamentBody(t, accept)))
		requests[index].Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		requests[index].Header.Set(echo.HeaderAuthorization, "Bearer "+organizer.AccessToken)
		responses[index] = httptest.NewRecorder()
		workers.Add(1)
		go func(index int) {
			defer workers.Done()
			server.ServeHTTP(responses[index], requests[index])
		}(index)
	}
	workers.Wait()
	statuses := map[int]int{}
	for index, response := range responses {
		statuses[response.Code]++
		validateIntegrationResponse(t, requests[index], response)
	}
	if statuses[200] != 1 || statuses[409] != 1 {
		t.Fatalf("competing reviews did not serialize capacity: %+v", statuses)
	}
	var count int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM applications WHERE tournament_id=$1 AND status='accepted'`, tournament.Id).Scan(&count); err != nil || count != 2 {
		t.Fatalf("accepted count=%d error=%v", count, err)
	}
}
