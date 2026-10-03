package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
)

type fakeDatabase struct{ err error }

func (db fakeDatabase) Ping(context.Context) error { return db.err }

func TestServerContractAndScaffold(t *testing.T) {
	server, err := NewServer(fakeDatabase{})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/health/live", "", 200},
		{"GET", "/health/ready", "", 200},
		{"GET", "/openapi.yaml", "", 200},
		{"GET", "/api/v1/tournaments", "", 501},
		{"GET", "/api/v1/tournaments?limit=101", "", 400},
		{"GET", "/api/v1/matches/not-a-uuid", "", 400},
		{"POST", "/api/v1/auth/register", `{"email":"invalid","password":"short","nickname":"a"}`, 400},
		{"POST", "/api/v1/auth/login", `{"email":"player@example.com","password":"some-password"}`, 501},
		{"GET", "/api/v1/users/me", "", 501},
		{"GET", "/api/v1/unknown", "", 404},
	} {
		t.Run(tc.method+tc.path, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			request.Header.Set("Authorization", "Bearer arbitrary-token")
			recorder := httptest.NewRecorder()
			server.ServeHTTP(recorder, request)
			if recorder.Code != tc.status {
				t.Fatalf("status %d, want %d: %s", recorder.Code, tc.status, recorder.Body.String())
			}
			if tc.status >= 400 {
				var body contract.Error
				if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Code == "" || body.Message == "" || body.RequestId == "" {
					t.Fatalf("incomplete error envelope: %+v", body)
				}
			}
		})
	}
}

func TestReadinessDetectsDatabaseFailure(t *testing.T) {
	server, err := NewServer(fakeDatabase{err: errors.New("database stopped")})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, httptest.NewRequest("GET", "/health/ready", nil))
	if recorder.Code != 503 {
		t.Fatalf("readiness: %d", recorder.Code)
	}
}
