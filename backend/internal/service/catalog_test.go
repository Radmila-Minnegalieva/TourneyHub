package service

import (
	"testing"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
)

func TestDisciplineValidation(t *testing.T) {
	for _, request := range []model.DisciplineRequest{
		{Name: " ", ScoreUnit: "points", MinRosterSize: 1, MaxRosterSize: 1},
		{Name: "Chess", ScoreUnit: " ", MinRosterSize: 1, MaxRosterSize: 1},
		{Name: "Chess", ScoreUnit: "points", MinRosterSize: 3, MaxRosterSize: 2},
	} {
		if validateDiscipline(request) == nil {
			t.Fatalf("invalid discipline accepted: %+v", request)
		}
	}
}
