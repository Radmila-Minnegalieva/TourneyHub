//go:build integration

package api

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/smtpclient"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository/postgres"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/service"
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func TestAdminRolesBlockingAndDisciplineIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	adminID, admin := seedIntegrationUser(t, db, "mainAdmin", contract.Participant, contract.Admin)
	userID, user := seedIntegrationUser(t, db, "managedUser", contract.Participant)
	adminPath, userPath := BaseURL+"/admin/users/"+adminID.String(), BaseURL+"/admin/users/"+userID.String()
	integrationRequest(t, server, "GET", BaseURL+"/admin/users", "", user.AccessToken, 403)
	integrationRequest(t, server, "GET", BaseURL+"/admin/users?search=managed", "", admin.AccessToken, 200)
	integrationRequest(t, server, "PUT", adminPath+"/block", `{"version":1,"blocked":true,"reason":"test"}`, admin.AccessToken, 409)
	integrationRequest(t, server, "PUT", adminPath+"/roles", `{"version":1,"roles":["participant"]}`, admin.AccessToken, 409)
	integrationRequest(t, server, "PUT", userPath+"/roles", `{"version":1,"roles":["organizer"]}`, admin.AccessToken, 422)
	profile := integrationDecode[contract.Profile](t, integrationRequest(t, server, "PUT", userPath+"/roles", `{"version":1,"roles":["participant","organizer","admin"]}`, admin.AccessToken, 200))
	if profile.Version != 2 || len(profile.Roles) != 3 {
		t.Fatalf("roles were not updated: %+v", profile)
	}
	integrationRequest(t, server, "PUT", userPath+"/block", `{"version":1,"blocked":true,"reason":"stale"}`, admin.AccessToken, 409)
	integrationRequest(t, server, "PUT", userPath+"/block", `{"version":2,"blocked":true}`, admin.AccessToken, 422)
	profile = integrationDecode[contract.Profile](t, integrationRequest(t, server, "PUT", userPath+"/block", `{"version":2,"blocked":true,"reason":"Rule violation"}`, admin.AccessToken, 200))
	if !profile.Blocked {
		t.Fatal("user was not blocked")
	}
	integrationRequest(t, server, "GET", BaseURL+"/users/me", "", user.AccessToken, 401)
	integrationRequest(t, server, "POST", BaseURL+"/auth/refresh", tournamentBody(t, contract.TokenRequest{Token: user.RefreshToken}), "", 401)
	integrationRequest(t, server, "PUT", adminPath+"/block", `{"version":1,"blocked":true,"reason":"last active admin"}`, admin.AccessToken, 409)
	profile = integrationDecode[contract.Profile](t, integrationRequest(t, server, "PUT", userPath+"/block", tournamentBody(t, contract.BlockUserRequest{Version: profile.Version, Blocked: false}), admin.AccessToken, 200))
	integrationRequest(t, server, "PUT", userPath+"/roles", tournamentBody(t, contract.UpdateRolesRequest{Version: profile.Version, Roles: []contract.Role{contract.Participant}}), admin.AccessToken, 200)
	integrationRequest(t, server, "POST", BaseURL+"/admin/disciplines", `{"name":"Bad roster","min_roster_size":3,"max_roster_size":1,"allows_draw":false,"score_unit":"points"}`, admin.AccessToken, 422)
	disciplineBody := `{"name":"Chess","min_roster_size":1,"max_roster_size":1,"allows_draw":true,"score_unit":"points"}`
	discipline := integrationDecode[contract.Discipline](t, integrationRequest(t, server, "POST", BaseURL+"/admin/disciplines", disciplineBody, admin.AccessToken, 201))
	integrationRequest(t, server, "POST", BaseURL+"/admin/disciplines", disciplineBody, admin.AccessToken, 409)
	disciplinePath := BaseURL + "/admin/disciplines/" + discipline.Id.String()
	integrationRequest(t, server, "PATCH", disciplinePath, `{"version":9,"active":false}`, admin.AccessToken, 409)
	integrationRequest(t, server, "PATCH", disciplinePath, `{"version":1,"min_roster_size":3}`, admin.AccessToken, 422)
	integrationRequest(t, server, "PATCH", disciplinePath, `{"version":1,"active":false}`, admin.AccessToken, 200)
	page := integrationDecode[contract.DisciplinePage](t, integrationRequest(t, server, "GET", BaseURL+"/disciplines", "", "", 200))
	if page.Total != 0 {
		t.Fatal("inactive discipline is still in the public catalog")
	}
	auditResponse := integrationRequest(t, server, "GET", BaseURL+"/admin/audit?entity_id="+userID.String(), "", admin.AccessToken, 200)
	if page := integrationDecode[contract.AuditEntryPage](t, auditResponse); page.Total < 4 {
		t.Fatalf("audit lost administrative changes: %+v", page)
	}
	if strings.Contains(auditResponse.Body.String(), "password_hash") || strings.Contains(auditResponse.Body.String(), user.AccessToken) {
		t.Fatal("audit leaked secrets")
	}
}

func TestModerationNotificationsExportAndReplayIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	_, admin := seedIntegrationUser(t, db, "moderator", contract.Participant, contract.Admin)
	_, organizer := seedIntegrationUser(t, db, "exportOrganizer", contract.Organizer)
	playerID, player := seedIntegrationUser(t, db, "exportPlayer", contract.Participant)
	_, outsider := seedIntegrationUser(t, db, "exportOutsider", contract.Participant)
	discipline := seedTournamentDiscipline(t, db)
	tournament := publishIntegrationTournament(t, server, createIntegrationTournament(t, server, discipline, organizer.AccessToken, 4), organizer.AccessToken)
	path := BaseURL + "/tournaments/" + tournament.Id.String()
	moderationPath := BaseURL + "/admin/tournaments/" + tournament.Id.String() + "/visibility"
	tournament = integrationDecode[contract.Tournament](t, integrationRequest(t, server, "PUT", moderationPath, tournamentBody(t, contract.ModerateTournamentRequest{Version: tournament.Version, Hidden: true, Reason: "Check the rules"}), admin.AccessToken, 200))
	integrationRequest(t, server, "GET", path, "", "", 404)
	adminPage := integrationDecode[contract.TournamentPage](t, integrationRequest(t, server, "GET", BaseURL+"/admin/tournaments", "", admin.AccessToken, 200))
	if adminPage.Total != 1 || !adminPage.Items[0].Hidden {
		t.Fatal("hidden tournament is absent from the administrative catalog")
	}
	notifications := integrationDecode[contract.NotificationPage](t, integrationRequest(t, server, "GET", BaseURL+"/users/me/notifications?unread=true", "", organizer.AccessToken, 200))
	if notifications.Total == 0 {
		t.Fatal("moderation did not notify the organizer")
	}
	readPath := BaseURL + "/users/me/notifications/" + notifications.Items[0].Id.String() + "/read"
	integrationRequest(t, server, "PUT", readPath, "", outsider.AccessToken, 404)
	integrationRequest(t, server, "PUT", readPath, "", organizer.AccessToken, 204)
	integrationRequest(t, server, "PUT", readPath, "", organizer.AccessToken, 204)
	tournament = integrationDecode[contract.Tournament](t, integrationRequest(t, server, "PUT", moderationPath, tournamentBody(t, contract.ModerateTournamentRequest{Version: tournament.Version, Hidden: false, Reason: "Rules verified"}), admin.AccessToken, 200))
	team := createApplicationTeam(t, server, "=SUM(1,2)", player.AccessToken)
	application := submitIntegrationApplication(t, server, tournament.Id, &team.Id, playerID, player.AccessToken)
	ctx := context.Background()
	if _, err := db.Exec(ctx, `UPDATE applications SET status='accepted',final_place=1 WHERE id=$1`, application.Id); err != nil {
		t.Fatal(err)
	}
	runnerUp := createApplicationTeam(t, server, "Runner Up", player.AccessToken)
	awayID, stageID, matchID := uuid.New(), uuid.New(), uuid.New()
	if _, err := db.Exec(ctx, `INSERT INTO applications(id,tournament_id,team_id,submitted_by,status,final_place) VALUES($1,$2,$3,$4,'accepted',2)`, awayID, tournament.Id, runnerUp.Id, playerID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO stages(id,tournament_id,kind,position,status) VALUES($1,$2,'playoffs',1,'completed')`, stageID, tournament.Id); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO matches(id,tournament_id,stage_id,round,position,home_application_id,away_application_id,winner_application_id,home_score,away_score,status) VALUES($1,$2,$3,1,1,$4,$5,$4,3,1,'completed')`, matchID, tournament.Id, stageID, application.Id, awayID); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `UPDATE tournaments SET status='completed' WHERE id=$1`, tournament.Id); err != nil {
		t.Fatal(err)
	}
	export := integrationRequest(t, server, "GET", path+"/export", "", "", 200)
	reader := csv.NewReader(strings.NewReader(strings.TrimPrefix(export.Body.String(), "\ufeff")))
	records, err := reader.ReadAll()
	if err != nil || len(records) != 4 || records[1][2] != "'=SUM(1,2)" || records[3][0] != "match" || records[3][3] != matchID.String() || records[3][6] != "3" || records[3][7] != "1" {
		t.Fatalf("unsafe or invalid CSV: %q, %v", export.Body.String(), err)
	}
	if err := postgres.New(db).Event(ctx, tournament.Id, "match_updated", map[string]any{"test": "first"}); err != nil {
		t.Fatal(err)
	}
	if err := postgres.New(db).Event(ctx, tournament.Id, "standings_updated", map[string]any{"test": "second"}); err != nil {
		t.Fatal(err)
	}
	api := New(integrationService(db))
	recorder := httptest.NewRecorder()
	c := server.NewContext(httptest.NewRequest("GET", path+"/events", nil), recorder)
	last, err := api.replayEvents(c, tournament.Id, 0)
	if err != nil || last == 0 || !strings.Contains(recorder.Body.String(), "event: match_updated") || !strings.Contains(recorder.Body.String(), "event: standings_updated") {
		t.Fatalf("events were not replayed: %s, %v", recorder.Body.String(), err)
	}
	recorder = httptest.NewRecorder()
	c = server.NewContext(httptest.NewRequest("GET", path+"/events", nil), recorder)
	if _, err := api.replayEvents(c, tournament.Id, last); err != nil || recorder.Body.Len() != 0 {
		t.Fatalf("events replayed twice: %q, %v", recorder.Body.String(), err)
	}
	recorder = httptest.NewRecorder()
	c = server.NewContext(httptest.NewRequest("GET", path+"/events", nil), recorder)
	if _, err := api.replayEvents(c, tournament.Id, last+1000); err != nil || !strings.Contains(recorder.Body.String(), "resync_required") {
		t.Fatalf("invalid cursor did not trigger resync: %q, %v", recorder.Body.String(), err)
	}
	// Административное удаление сохраняет историю и освобождает резерв игроков.
	var version int
	if err := db.QueryRow(ctx, `SELECT version FROM tournaments WHERE id=$1`, tournament.Id).Scan(&version); err != nil {
		t.Fatal(err)
	}
	integrationRequest(t, server, "DELETE", BaseURL+"/admin/tournaments/"+tournament.Id.String(), tournamentBody(t, contract.ReasonRequest{Version: version, Reason: "Delete demonstration tournament"}), admin.AccessToken, 204)
	integrationRequest(t, server, "GET", path, "", "", 404)
	var claims int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM tournament_player_claims WHERE tournament_id=$1`, tournament.Id).Scan(&claims); err != nil || claims != 0 {
		t.Fatalf("deleted tournament retained claims: %d, %v", claims, err)
	}
	var deletionNotifications int
	if err := db.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE resource_id=$1 AND kind='tournament_deleted'`, tournament.Id).Scan(&deletionNotifications); err != nil || deletionNotifications != 1 {
		t.Fatalf("delete did not notify organizer: count=%d, err=%v", deletionNotifications, err)
	}
}

func ancillaryUpload(t *testing.T, server *echo.Echo, filename, purpose string, data []byte, token string, want int) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("purpose", purpose); err != nil {
		t.Fatal(err)
	}
	file, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("POST", BaseURL+"/files", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != want {
		t.Fatalf("upload status=%d, want=%d: %s", recorder.Code, want, recorder.Body.String())
	}
	validateIntegrationResponse(t, request, recorder)
	return recorder
}

func TestFilesPrivacyAndSignatureIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	_, owner := seedIntegrationUser(t, db, "fileOwner", contract.Participant)
	_, stranger := seedIntegrationUser(t, db, "fileStranger", contract.Participant)
	_, admin := seedIntegrationUser(t, db, "fileAdmin", contract.Participant, contract.Admin)
	image := testPNG(t)
	avatar := integrationDecode[contract.File](t, ancillaryUpload(t, server, "../../avatar.png", "avatar", image, owner.AccessToken, 201))
	if avatar.Filename != "avatar.png" || avatar.ContentType != contract.Imagepng {
		t.Fatalf("invalid avatar metadata: %+v", avatar)
	}
	response := integrationRequest(t, server, "GET", BaseURL+"/files/"+avatar.Id.String(), "", "", 200)
	if !bytes.Equal(response.Body.Bytes(), image) || !strings.HasPrefix(response.Header().Get("Content-Type"), "image/png") || response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("download changed file contents or omitted defensive headers")
	}
	pdf := []byte("%PDF-1.4\n%%EOF\n")
	protocol := integrationDecode[contract.File](t, ancillaryUpload(t, server, "result.pdf", "match_protocol", pdf, owner.AccessToken, 201))
	filePath := BaseURL + "/files/" + protocol.Id.String()
	integrationRequest(t, server, "GET", filePath, "", "", 403)
	integrationRequest(t, server, "GET", filePath, "", stranger.AccessToken, 403)
	integrationRequest(t, server, "GET", filePath, "", admin.AccessToken, 200)
	if response := integrationRequest(t, server, "GET", filePath, "", owner.AccessToken, 200); !bytes.Equal(response.Body.Bytes(), pdf) {
		t.Fatal("protocol download changed the stored content")
	}
	ancillaryUpload(t, server, "fake.png", "avatar", []byte("<html>not an image</html>"), owner.AccessToken, 422)
	ancillaryUpload(t, server, "forbidden.pdf", "avatar", pdf, owner.AccessToken, 422)
	ancillaryUpload(t, server, "large.pdf", "match_protocol", make([]byte, service.MaxFileBytes+1), owner.AccessToken, 422)
}

func TestAttachedProtocolAccessIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	_, organizer := seedIntegrationUser(t, db, "protocolOrganizer", contract.Organizer)
	homeID, home := seedIntegrationUser(t, db, "protocolHome", contract.Participant)
	awayID, away := seedIntegrationUser(t, db, "protocolAway", contract.Participant)
	newCaptainID, newCaptain := seedIntegrationUser(t, db, "protocolCaptain", contract.Participant)
	refereeID, referee := seedIntegrationUser(t, db, "protocolReferee", contract.Referee)
	_, outsider := seedIntegrationUser(t, db, "protocolOutsider", contract.Participant)
	tournament := publishIntegrationTournament(t, server, createIntegrationTournament(t, server, seedTournamentDiscipline(t, db), organizer.AccessToken, 4), organizer.AccessToken)
	team := createApplicationTeam(t, server, "Protocol team", home.AccessToken)
	homeApplication := submitIntegrationApplication(t, server, tournament.Id, &team.Id, homeID, home.AccessToken)
	awayApplication := submitIntegrationApplication(t, server, tournament.Id, nil, awayID, away.AccessToken)
	protocol := integrationDecode[contract.File](t, ancillaryUpload(t, server, "signed-result.pdf", "match_protocol", []byte("%PDF-1.4\n%%EOF\n"), home.AccessToken, 201))
	ctx := context.Background()
	stageID, matchID, reportID := uuid.New(), uuid.New(), uuid.New()
	for _, statement := range []struct {
		sql  string
		args []any
	}{
		{`INSERT INTO stages(id,tournament_id,kind,position) VALUES($1,$2,'playoffs',1)`, []any{stageID, tournament.Id}},
		{`INSERT INTO matches(id,tournament_id,stage_id,round,position,home_application_id,away_application_id,referee_id) VALUES($1,$2,$3,1,1,$4,$5,$6)`, []any{matchID, tournament.Id, stageID, homeApplication.Id, awayApplication.Id, refereeID}},
		{`INSERT INTO result_reports(id,match_id,author_id,source,home_score,away_score) VALUES($1,$2,$3,'captain',1,0)`, []any{reportID, matchID, homeID}},
		{`INSERT INTO report_files(report_id,file_id) VALUES($1,$2)`, []any{reportID, protocol.Id}},
		{`INSERT INTO team_members(team_id,user_id) VALUES($1,$2)`, []any{team.Id, newCaptainID}},
	} {
		if _, err := db.Exec(ctx, statement.sql, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	// Новый капитан не входит в снимок игроков матча, но получает доступ к протоколу команды.
	integrationRequest(t, server, "PUT", BaseURL+"/teams/"+team.Id.String()+"/captain", tournamentBody(t, contract.TransferCaptainRequest{Version: team.Version, UserId: newCaptainID}), home.AccessToken, 200)
	filePath := BaseURL + "/files/" + protocol.Id.String()
	for _, token := range []string{home.AccessToken, away.AccessToken, newCaptain.AccessToken, organizer.AccessToken, referee.AccessToken} {
		integrationRequest(t, server, "GET", filePath, "", token, 200)
	}
	integrationRequest(t, server, "GET", filePath, "", outsider.AccessToken, 403)
	integrationRequest(t, server, "GET", filePath, "", "", 403)
}

func TestSMTPOutboxRetryAndRedactionIntegration(t *testing.T) {
	_, db := newIntegrationServer(t)
	cfg := config.Default()
	repo := postgres.New(db)
	svc := service.New(repo, cfg, nil)
	ctx := context.Background()
	if err := postgres.New(db).Outbox(ctx, "email", uuid.New(), map[string]any{"to": "recipient@example.com", "subject": "TourneyHub notification", "text": "Secret verification text"}); err != nil {
		t.Fatal(err)
	}

	if err := svc.DeliverMail(ctx); err != nil {
		t.Fatal(err)
	}
	cfg.SMTP.Address = "127.0.0.1:0"
	cfg.SMTP.From = "sender@example.com"
	cfg.SMTP.AllowInsecure = true
	svc = service.New(repo, cfg, smtpclient.New(cfg.SMTP))
	if err := svc.DeliverMail(ctx); err != nil {
		t.Fatal(err)
	}
	var attempts int
	var deferred, delivered bool
	if err := db.QueryRow(ctx, `SELECT attempts,available_at>now(),delivered_at IS NOT NULL FROM outbox_events`).Scan(&attempts, &deferred, &delivered); err != nil || attempts != 1 || !deferred || delivered {
		t.Fatalf("failed delivery was not rescheduled: attempts=%d deferred=%v delivered=%v error=%v", attempts, deferred, delivered, err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	mailResult := make(chan string, 1)
	go serveTestSMTP(listener, mailResult)
	cfg.SMTP.Address = listener.Addr().String()
	svc = service.New(repo, cfg, smtpclient.New(cfg.SMTP))
	if _, err := db.Exec(ctx, `UPDATE outbox_events SET available_at=now()`); err != nil {
		t.Fatal(err)
	}
	if err := svc.DeliverMail(ctx); err != nil {
		t.Fatal(err)
	}
	var redacted bool
	if err := db.QueryRow(ctx, `SELECT attempts,delivered_at IS NOT NULL,payload='{"redacted":true}'::jsonb FROM outbox_events`).Scan(&attempts, &delivered, &redacted); err != nil || attempts != 2 || !delivered || !redacted {
		t.Fatalf("delivery state or redaction is incorrect: attempts=%d delivered=%v redacted=%v error=%v", attempts, delivered, redacted, err)
	}
	select {
	case message := <-mailResult:
		if !strings.Contains(message, "Secret verification text") || !strings.Contains(message, "To: recipient@example.com") {
			t.Fatalf("SMTP received the wrong message: %s", message)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("SMTP receiver did not finish")
	}
}

func serveTestSMTP(listener net.Listener, result chan<- string) {
	conn, err := listener.Accept()
	if err != nil {
		result <- "accept: " + err.Error()
		return
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	client := textproto.NewConn(conn)
	_ = client.PrintfLine("220 localhost test SMTP")
	var message string
	for {
		line, err := client.ReadLine()
		if err != nil {
			result <- message + "\nread: " + err.Error()
			return
		}
		switch {
		case strings.HasPrefix(line, "EHLO"), strings.HasPrefix(line, "HELO"), strings.HasPrefix(line, "MAIL"), strings.HasPrefix(line, "RCPT"):
			_ = client.PrintfLine("250 OK")
		case line == "DATA":
			_ = client.PrintfLine("354 Send mail")
			data, err := io.ReadAll(client.DotReader())
			if err != nil {
				result <- fmt.Sprint(err)
				return
			}
			message = string(data)
			_ = client.PrintfLine("250 Accepted")
		case line == "QUIT":
			_ = client.PrintfLine("221 Bye")
			result <- message
			return
		default:
			_ = client.PrintfLine("500 Unsupported")
		}
	}
}
