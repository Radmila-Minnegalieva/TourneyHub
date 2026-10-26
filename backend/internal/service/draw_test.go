package service

import (
	"testing"

	model "github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"

	"github.com/google/uuid"
)

func TestGroupsMayQualifyAllParticipants(t *testing.T) {
	tournament := model.Tournament{Id: uuid.New(), Format: model.TournamentFormatGroupsPlayoffs, Rules: model.Rules{RoundRobinLegs: 1, GroupSettings: &model.GroupSettings{GroupCount: 2, QualifiersPerGroup: 2}}}
	participants := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	preview, err := buildDraw(tournament, participants)
	if err != nil {
		t.Fatal(err)
	}
	if len(preview.Groups) != 2 || len(preview.Matches) != 2 || len(preview.Stages) != 2 {
		t.Fatalf("unexpected group draw: %+v", preview)
	}
	tournament.Rules.GroupSettings.QualifiersPerGroup = 3
	if _, err = buildDraw(tournament, participants); err == nil {
		t.Fatal("accepted more qualifiers than group participants")
	}
}
