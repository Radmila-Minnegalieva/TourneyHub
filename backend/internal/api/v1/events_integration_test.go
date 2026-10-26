//go:build integration

package api

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository/postgres"
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
)

func TestLiveSSEUpdatesAndVisibilityIntegration(t *testing.T) {
	server, db := newIntegrationServer(t)
	_, organizer := seedIntegrationUser(t, db, "sseOrganizer", contract.Organizer)
	tournament := publishIntegrationTournament(t, server, createIntegrationTournament(t, server, seedTournamentDiscipline(t, db), organizer.AccessToken, 4), organizer.AccessToken)
	path := BaseURL + "/tournaments/" + tournament.Id.String() + "/events"
	bad := httptest.NewRequest("GET", path, nil)
	bad.Header.Set("Last-Event-ID", "invalid")
	rec := httptest.NewRecorder()
	server.ServeHTTP(rec, bad)
	if rec.Code != 400 {
		t.Fatalf("invalid SSE cursor: %d", rec.Code)
	}
	httpServer := httptest.NewServer(server)
	defer httpServer.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", httpServer.URL+path, nil)
	if err != nil {
		t.Fatal(err)
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 || response.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("SSE headers: %d %v", response.StatusCode, response.Header)
	}
	reader := bufio.NewReader(response.Body)
	line, err := reader.ReadString('\n')
	if err != nil || !strings.Contains(line, "connected") {
		t.Fatalf("stream did not connect: %q %v", line, err)
	}
	if err := postgres.New(db).Event(ctx, tournament.Id, "match_updated", map[string]any{"live": true}); err != nil {
		t.Fatal(err)
	}
	found := false
	for !found {
		line, err = reader.ReadString('\n')
		if err != nil {
			t.Fatalf("stream missed committed event: %v", err)
		}
		found = strings.Contains(line, `"live": true`) || strings.Contains(line, `"live":true`)
	}
	if _, err := db.Exec(ctx, `UPDATE tournaments SET hidden=true WHERE id=$1`, tournament.Id); err != nil {
		t.Fatal(err)
	}
	if _, err := io.ReadAll(reader); err != nil {
		t.Fatalf("hidden tournament did not close SSE cleanly: %v", err)
	}
}
