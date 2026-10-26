//go:build integration

package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/config"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/security"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/migration"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository/postgres"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/service"
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// newIntegrationServer создаёт изолированную схему в отдельной тестовой БД.
// Ни public, ни данные другого теста не удаляются даже при ошибке сценария.
func newIntegrationServer(t *testing.T) (*echo.Echo, *pgxpool.Pool) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL is required for integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if _, err = admin.Exec(ctx, `CREATE SCHEMA `+pgx.Identifier{schema}.Sanitize()); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), 10*time.Second)
		defer stop()
		if _, err := admin.Exec(cleanup, `DROP SCHEMA `+pgx.Identifier{schema}.Sanitize()+` CASCADE`); err != nil {
			t.Errorf("drop isolated test schema: %v", err)
		}
		admin.Close()
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", schema)
	u.RawQuery = query.Encode()
	provider, migrationDB, err := migration.Open(ctx, u.String(), filepath.Join("..", "..", "..", "migrations"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		migrationDB.Close()
		t.Fatal(err)
	}
	if err := migrationDB.Close(); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	server, err := NewServer(pool, integrationService(pool), config.Default().HTTP)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = server.Close() })
	return server, pool
}

func integrationRequest(t *testing.T, server *echo.Echo, method, path, body, token string, want int) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	if token != "" {
		request.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != want {
		t.Fatalf("%s %s: status=%d, want=%d, body=%s", method, path, recorder.Code, want, recorder.Body.String())
	}
	validateIntegrationResponse(t, request, recorder)
	return recorder
}

func validateIntegrationResponse(t *testing.T, request *http.Request, recorder *httptest.ResponseRecorder) {
	t.Helper()
	spec, err := contract.GetSwagger()
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(spec)
	if err != nil {
		t.Fatal(err)
	}
	route, pathParams, err := router.FindRoute(request)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{Request: request, Route: route, PathParams: pathParams},
		Status:                 recorder.Code,
		Header:                 recorder.Header(),
		Body:                   io.NopCloser(bytes.NewReader(recorder.Body.Bytes())),
		Options:                &openapi3filter.Options{IncludeResponseStatus: true},
	}
	if err := openapi3filter.ValidateResponse(context.Background(), input); err != nil {
		t.Fatalf("response violates OpenAPI (%s %s): %v; body=%s", request.Method, request.URL.Path, err, recorder.Body.String())
	}
}

// seedIntegrationUser создаёт проверенного пользователя и обычные DB-сессии.
// Используется для проверки защищённых бизнес-сценариев без повторения auth-сценария.
func seedIntegrationUser(t *testing.T, pool *pgxpool.Pool, nickname string, roles ...contract.Role) (uuid.UUID, contract.Tokens) {
	t.Helper()
	ctx := context.Background()
	hash, err := security.HashPassword("test-password-123")
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	var id uuid.UUID
	err = tx.QueryRow(ctx, `INSERT INTO users(email,password_hash,nickname,personal_data_consent_at,email_verified_at) VALUES($1,$2,$3,now(),now()) RETURNING id`, nickname+"@example.com", hash, nickname).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) == 0 {
		roles = []contract.Role{contract.Participant}
	}
	for _, role := range roles {
		if _, err := tx.Exec(ctx, `INSERT INTO user_roles(user_id,role) VALUES($1,$2)`, id, string(role)); err != nil {
			t.Fatal(err)
		}
	}
	tokens, err := seedSession(ctx, tx, id)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	return id, tokens
}

func integrationDecode[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()
	var value T
	if err := json.Unmarshal(recorder.Body.Bytes(), &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func TestIdentityLifecycleIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	ctx := context.Background()
	const registration = `{"email":"new@example.com","nickname":"new_player","password":"password-123456","personal_data_consent":true}`
	integrationRequest(t, server, "POST", BaseURL+"/auth/register", registration, "", 202)
	integrationRequest(t, server, "POST", BaseURL+"/auth/register", registration, "", 409)
	var id uuid.UUID
	var verification string
	if err := db.QueryRow(ctx, `SELECT u.id,o.payload->>'token' FROM users u JOIN outbox_events o ON o.aggregate_id=u.id WHERE u.email='new@example.com' AND o.topic='email'`).Scan(&id, &verification); err != nil {
		t.Fatal(err)
	}
	const login = `{"email":"new@example.com","password":"password-123456"}`
	integrationRequest(t, server, "POST", BaseURL+"/auth/login", login, "", 401)
	verificationBody := `{"token":"` + verification + `"}`
	integrationRequest(t, server, "POST", BaseURL+"/auth/verify-email", verificationBody, "", 204)
	integrationRequest(t, server, "POST", BaseURL+"/auth/verify-email", verificationBody, "", 422)
	tokens := integrationDecode[contract.Tokens](t, integrationRequest(t, server, "POST", BaseURL+"/auth/login", login, "", 200))
	p := integrationDecode[contract.Profile](t, integrationRequest(t, server, "GET", BaseURL+"/users/me", "", tokens.AccessToken, 200))
	if p.Id != id || !p.EmailVerified || len(p.Roles) != 1 || p.Roles[0] != contract.Participant {
		t.Fatalf("unexpected profile: %+v", p)
	}
	integrationRequest(t, server, "PATCH", BaseURL+"/users/me", `{"version":2,"timezone":"Bad/Timezone"}`, tokens.AccessToken, 422)
	integrationRequest(t, server, "PATCH", BaseURL+"/users/me", `{"version":2,"timezone":"Europe/Moscow","name":"Игрок"}`, tokens.AccessToken, 200)
	integrationRequest(t, server, "PATCH", BaseURL+"/users/me", `{"version":2,"name":"stale"}`, tokens.AccessToken, 409)
	refreshed := integrationDecode[contract.Tokens](t, integrationRequest(t, server, "POST", BaseURL+"/auth/refresh", `{"token":"`+tokens.RefreshToken+`"}`, "", 200))
	integrationRequest(t, server, "GET", BaseURL+"/users/me", "", tokens.AccessToken, 401)
	integrationRequest(t, server, "POST", BaseURL+"/auth/refresh", `{"token":"`+tokens.RefreshToken+`"}`, "", 401)
	integrationRequest(t, server, "POST", BaseURL+"/auth/logout", `{"token":"`+refreshed.RefreshToken+`"}`, "", 204)
	integrationRequest(t, server, "POST", BaseURL+"/auth/logout", `{"token":"`+refreshed.RefreshToken+`"}`, "", 204)
	integrationRequest(t, server, "GET", BaseURL+"/users/me", "", refreshed.AccessToken, 401)
	tokens = integrationDecode[contract.Tokens](t, integrationRequest(t, server, "POST", BaseURL+"/auth/login", login, "", 200))
	secondSession := additionalIntegrationSession(t, db, id)
	integrationRequest(t, server, "POST", BaseURL+"/auth/password-reset", `{"email":"unknown@example.com"}`, "", 202)
	integrationRequest(t, server, "POST", BaseURL+"/auth/password-reset", `{"email":"new@example.com"}`, "", 202)
	var resetToken string
	if err := db.QueryRow(ctx, `SELECT payload->>'token' FROM outbox_events WHERE aggregate_id=$1 AND payload->>'purpose'='reset_password'`, id).Scan(&resetToken); err != nil {
		t.Fatal(err)
	}
	resetBody := `{"token":"` + resetToken + `","password":"new-password-123456"}`
	integrationRequest(t, server, "POST", BaseURL+"/auth/password-reset/confirm", resetBody, "", 204)
	integrationRequest(t, server, "POST", BaseURL+"/auth/password-reset/confirm", resetBody, "", 422)
	integrationRequest(t, server, "GET", BaseURL+"/users/me", "", tokens.AccessToken, 401)
	integrationRequest(t, server, "POST", BaseURL+"/auth/refresh", `{"token":"`+tokens.RefreshToken+`"}`, "", 401)
	integrationRequest(t, server, "GET", BaseURL+"/users/me", "", secondSession.AccessToken, 401)
	integrationRequest(t, server, "POST", BaseURL+"/auth/refresh", `{"token":"`+secondSession.RefreshToken+`"}`, "", 401)
	integrationRequest(t, server, "POST", BaseURL+"/auth/login", login, "", 401)
	integrationRequest(t, server, "POST", BaseURL+"/auth/login", `{"email":"new@example.com","password":"new-password-123456"}`, "", 200)
	// Шестая попытка в одном окне отвергается независимо от корректности пароля.
	integrationRequest(t, server, "POST", BaseURL+"/auth/login", login, "", 429)
}

func additionalIntegrationSession(t *testing.T, db *pgxpool.Pool, id uuid.UUID) contract.Tokens {
	t.Helper()
	tx, err := db.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	tokens, err := seedSession(context.Background(), tx, id)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(context.Background()); err != nil {
		t.Fatal(err)
	}
	return tokens
}

func TestConcurrentRefreshIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	id, initial := seedIntegrationUser(t, db, "parallel_user")
	const workers = 8
	recorders := make(chan *httptest.ResponseRecorder, workers)
	var group sync.WaitGroup
	for range workers {
		group.Go(func() {
			request := httptest.NewRequest("POST", BaseURL+"/auth/refresh", strings.NewReader(`{"token":"`+initial.RefreshToken+`"}`))
			request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, request)
			recorders <- recorder
		})
	}
	group.Wait()
	close(recorders)
	successes := 0
	for recorder := range recorders {
		switch recorder.Code {
		case http.StatusOK:
			successes++
		case http.StatusUnauthorized:
		default:
			t.Fatalf("concurrent refresh status=%d body=%s", recorder.Code, recorder.Body.String())
		}
	}
	if successes != 1 {
		t.Fatalf("a refresh token must be consumed once, successes=%d", successes)
	}
	var activeSessions int
	if err := db.QueryRow(context.Background(), `SELECT count(*) FROM refresh_sessions WHERE user_id=$1 AND revoked_at IS NULL`, id).Scan(&activeSessions); err != nil {
		t.Fatal(err)
	}
	if activeSessions != 1 {
		t.Fatalf("active sessions=%d, want=1", activeSessions)
	}
}

func TestLoginRateIgnoresForwardedIPIntegration(t *testing.T) {
	server, _ := newIntegrationServer(t)
	for attempt := 1; attempt <= 6; attempt++ {
		request := httptest.NewRequest("POST", BaseURL+"/auth/login", strings.NewReader(`{"email":"missing@example.com","password":"test-password-123"}`))
		request.RemoteAddr = "192.0.2.12:4321"
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		request.Header.Set(echo.HeaderXForwardedFor, "198.51.100."+strconv.Itoa(attempt))
		recorder := httptest.NewRecorder()
		server.ServeHTTP(recorder, request)
		want := http.StatusUnauthorized
		if attempt == 6 {
			want = http.StatusTooManyRequests
		}
		if recorder.Code != want {
			t.Fatalf("attempt=%d status=%d, want=%d body=%s", attempt, recorder.Code, want, recorder.Body.String())
		}
		validateIntegrationResponse(t, request, recorder)
	}
}

func TestBlockedAccountAndUnicodePasswordIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	_, admin := seedIntegrationUser(t, db, "admin_user", contract.Participant, contract.Admin)
	id, account := seedIntegrationUser(t, db, "blocked_user")
	path := BaseURL + "/admin/users/" + id.String() + "/block"
	integrationRequest(t, server, "PUT", path, `{"version":1,"blocked":true,"reason":"Moderation"}`, admin.AccessToken, 200)
	integrationRequest(t, server, "GET", BaseURL+"/users/me", "", account.AccessToken, 401)
	integrationRequest(t, server, "POST", BaseURL+"/auth/refresh", `{"token":"`+account.RefreshToken+`"}`, "", 401)
	integrationRequest(t, server, "PUT", path, `{"version":2,"blocked":false}`, admin.AccessToken, 200)
	integrationRequest(t, server, "GET", BaseURL+"/users/me", "", account.AccessToken, 401)
	unicodePassword := strings.Repeat("я", 36)
	hash, err := security.HashPassword(unicodePassword)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(context.Background(), `UPDATE users SET password_hash=$2 WHERE id=$1`, id, hash); err != nil {
		t.Fatal(err)
	}
	integrationRequest(t, server, "POST", BaseURL+"/auth/login", `{"email":"blocked_user@example.com","password":"`+unicodePassword+`x"}`, "", 401)
	integrationRequest(t, server, "POST", BaseURL+"/auth/login", `{"email":"blocked_user@example.com","password":"`+unicodePassword+`"}`, "", 200)
}

func TestTeamsInvitationsAndOwnershipIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	_, captain := seedIntegrationUser(t, db, "captain")
	memberID, member := seedIntegrationUser(t, db, "member")
	_, outsider := seedIntegrationUser(t, db, "outsider")
	team := integrationDecode[contract.Team](t, integrationRequest(t, server, "POST", BaseURL+"/teams", `{"name":"Students"}`, captain.AccessToken, 201))
	path := BaseURL + "/teams/" + team.Id.String()
	integrationRequest(t, server, "PATCH", path, `{"version":1,"name":"Intrusion"}`, outsider.AccessToken, 403)
	link := integrationDecode[contract.InvitationLink](t, integrationRequest(t, server, "POST", path+"/invitations", `{"nickname":"member"}`, captain.AccessToken, 201))
	decisionPath := BaseURL + "/invitations/" + link.Invitation.Id.String() + "/decision"
	integrationRequest(t, server, "POST", decisionPath, `{"decision":"accept"}`, outsider.AccessToken, 403)
	integrationRequest(t, server, "POST", decisionPath, `{"decision":"accept"}`, member.AccessToken, 200)
	integrationRequest(t, server, "POST", decisionPath, `{"decision":"accept"}`, member.AccessToken, 409)
	public := integrationRequest(t, server, "GET", path+"/members", "", "", 200)
	if strings.Contains(public.Body.String(), "email") || strings.Contains(public.Body.String(), "roles") {
		t.Fatal("public membership leaked private user fields")
	}
	items := integrationDecode[[]contract.TeamMember](t, public)
	if len(items) != 2 {
		t.Fatalf("members=%d, want=2", len(items))
	}
	integrationRequest(t, server, "PUT", path+"/captain", `{"version":1,"user_id":"`+memberID.String()+`"}`, captain.AccessToken, 409)
	integrationRequest(t, server, "PUT", path+"/captain", `{"version":2,"user_id":"`+memberID.String()+`"}`, captain.AccessToken, 200)
	integrationRequest(t, server, "PATCH", path, `{"version":3,"name":"Denied"}`, captain.AccessToken, 403)
	integrationRequest(t, server, "PATCH", path, `{"version":3,"description":"Сборная университета"}`, member.AccessToken, 200)
	link = integrationDecode[contract.InvitationLink](t, integrationRequest(t, server, "POST", path+"/invitations", `{}`, member.AccessToken, 201))
	decisionPath = BaseURL + "/invitations/" + link.Invitation.Id.String() + "/decision"
	integrationRequest(t, server, "POST", decisionPath, `{"decision":"accept","token":"wrong"}`, outsider.AccessToken, 403)
	integrationRequest(t, server, "POST", decisionPath, `{"decision":"accept","token":"`+link.Token+`"}`, outsider.AccessToken, 200)
	team = integrationDecode[contract.Team](t, integrationRequest(t, server, "GET", path, "", "", 200))
	deleteBody := `{"version":` + stringMustJSON(t, team.Version) + `}`
	integrationRequest(t, server, "DELETE", path, deleteBody, captain.AccessToken, 403)
	integrationRequest(t, server, "DELETE", path, deleteBody, member.AccessToken, 204)
	integrationRequest(t, server, "GET", path, "", "", 404)
}

func stringMustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// seedSession создаёт реальные записи сессии внутри транзакции тестового fixture.
func seedSession(ctx context.Context, tx pgx.Tx, userID uuid.UUID) (contract.Tokens, error) {
	var tokens contract.Tokens
	access, err := security.NewToken()
	if err != nil {
		return tokens, err
	}
	refresh, err := security.NewToken()
	if err != nil {
		return tokens, err
	}
	var id uuid.UUID
	if err := tx.QueryRow(ctx, `INSERT INTO refresh_sessions(user_id,token_hash,expires_at) VALUES($1,$2,now()+interval '30 days') RETURNING id`, userID, security.TokenHash(refresh)).Scan(&id); err != nil {
		return tokens, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO access_sessions(user_id,refresh_session_id,token_hash,expires_at) VALUES($1,$2,$3,now()+interval '15 minutes')`, userID, id, security.TokenHash(access)); err != nil {
		return tokens, err
	}
	return contract.Tokens{AccessToken: access, RefreshToken: refresh, ExpiresIn: 900, TokenType: "Bearer"}, nil
}

func integrationService(pool *pgxpool.Pool) *service.Service {
	return service.New(postgres.New(pool), config.Default(), nil)
}
