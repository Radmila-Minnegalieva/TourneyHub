// File rating.go implements the Elo value and a documented composition example.
// It is linked from the generated project documentation.
package model

import (
	"errors"
	"math"
)

// InitialRating is the Elo value assigned to a team before its first match.
const InitialRating = 1500.0

// RatingK is the adjustment coefficient from requirement FR-35.
const RatingK = 32.0

// EloRating is a team's rating in one discipline.
//
// Ratings start at InitialRating. Each confirmed match changes the value using
// coefficient RatingK. Values stay unrounded in storage; the interface may round
// them for display. A bye does not constitute a rated match.
type EloRating struct {
	// TeamID identifies the team whose rating is being calculated.
	TeamID string
	// DisciplineID scopes the rating to a single competition discipline.
	DisciplineID string
	// Value is the current, unrounded Elo value.
	Value float64
}

// MatchRating describes the rating after a confirmed match.
//
// It embeds EloRating to reuse the team, discipline and resulting value.
// This is Go composition, the analogue used here for the assignment's derived type.
type MatchRating struct {
	// EloRating contains the resulting team rating and its identifiers.
	EloRating
	// Before is the value before the match.
	Before float64
	// Delta is the signed adjustment: Value minus Before.
	Delta float64
}

// NewEloRating returns an initial rating for a team in a discipline.
//
// teamID and disciplineID are domain identifiers supplied by the caller.
// The returned EloRating has Value equal to 1500 and does not perform persistence.
func NewEloRating(teamID, disciplineID string) EloRating {
	return EloRating{TeamID: teamID, DisciplineID: disciplineID, Value: InitialRating}
}

// AfterMatch calculates a new rating without changing the receiver.
//
// opponent is the opponent's unrounded rating before the match. score must be
// 1 for a win, 0.5 for a draw, or 0 for a loss. Both ratings must be finite.
// The returned MatchRating contains the old value, signed delta and new value.
// An invalid score or non-finite rating returns a zero MatchRating and an error.
//
// The caller calculates both opponents from their original ratings and persists
// the result in the same transaction as the confirmed match. Their adjustments
// are equal and opposite. The method does not validate tournament draw rules.
func (r EloRating) AfterMatch(opponent, score float64) (MatchRating, error) {
	if math.IsNaN(r.Value) || math.IsInf(r.Value, 0) || math.IsNaN(opponent) || math.IsInf(opponent, 0) {
		return MatchRating{}, errors.New("ratings must be finite")
	}
	if score != 0 && score != 0.5 && score != 1 {
		return MatchRating{}, errors.New("score must be 0, 0.5 or 1")
	}
	expected := 1 / (1 + math.Pow(10, (opponent-r.Value)/400))
	delta := RatingK * (score - expected)
	result := r
	result.Value += delta
	return MatchRating{EloRating: result, Before: r.Value, Delta: delta}, nil
}
