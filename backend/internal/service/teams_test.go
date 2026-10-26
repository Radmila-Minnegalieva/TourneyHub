package service

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/security"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"

	"github.com/google/uuid"
)

func TestTeamVersion(t *testing.T) {
	team := model.Team{Version: 3}
	if err := teamVersion(team, 3); err != nil {
		t.Fatal(err)
	}
	if err := teamVersion(team, 2); err == nil {
		t.Fatal("stale version accepted")
	}
}

func TestInvitationDecisionPolicy(t *testing.T) {
	actor, other := uuid.New(), uuid.New()
	now := time.Date(2026, 10, 26, 10, 0, 0, 0, time.UTC)
	secret, wrong := "one-time-token", "wrong-token"
	for _, tc := range []struct {
		name     string
		invitee  *uuid.UUID
		token    *string
		status   model.InvitationStatus
		expires  time.Time
		personal bool
		want     apperror.Kind
	}{
		{"addressed", &actor, nil, model.InvitationStatusPending, now.Add(time.Hour), false, ""},
		{"other player", &other, &secret, model.InvitationStatusPending, now.Add(time.Hour), false, apperror.Forbidden},
		{"link", nil, &secret, model.InvitationStatusPending, now.Add(time.Hour), false, ""},
		{"missing secret", nil, nil, model.InvitationStatusPending, now.Add(time.Hour), false, apperror.Forbidden},
		{"wrong secret", nil, &wrong, model.InvitationStatusPending, now.Add(time.Hour), false, apperror.Forbidden},
		{"expired", &actor, nil, model.InvitationStatusPending, now, false, apperror.Conflict},
		{"already accepted", &actor, nil, model.InvitationStatusAccepted, now.Add(time.Hour), false, apperror.Conflict},
		{"personal team", &actor, nil, model.InvitationStatusPending, now.Add(time.Hour), true, apperror.Conflict},
	} {
		t.Run(tc.name, func(t *testing.T) {
			state := lockedInvitation{invitation: model.Invitation{InviteeId: tc.invitee, Status: tc.status, ExpiresAt: tc.expires}, team: model.Team{Personal: tc.personal}, tokenHash: security.TokenHash(secret)}
			err := state.validate(actor, tc.token, now)
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var applicationError *apperror.Error
			if !errors.As(err, &applicationError) || applicationError.Kind != tc.want {
				t.Fatalf("error=%v, want kind=%s", err, tc.want)
			}
		})
	}
}

func TestInvitationURL(t *testing.T) {
	const webURL = "https://tourney.example/student/"
	id := uuid.New()
	token := "secret+token=&"
	u, err := url.Parse(invitationURL(webURL, id, token))
	if err != nil {
		t.Fatal(err)
	}
	if u.Scheme != "https" || u.Host != "tourney.example" || u.Path != "/student/invitations/"+id.String() || u.Query().Get("token") != token {
		t.Fatalf("invalid invitation URL: %s", u)
	}
}
