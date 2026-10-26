package service

import (
	"context"
	"math"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

type ratingMatch struct {
	id         uuid.UUID
	tournament uuid.UUID
	home       uuid.UUID
	away       uuid.UUID
	homeScore  int
	awayScore  int
}

type ratingValue struct {
	value  float64
	played int
}

// Исправление старого результата воспроизводит последующие матчи по времени
// первоначальной фиксации; прежние ревизии истории сохраняются.
func rebuildRatings(ctx context.Context, q repository.Store, discipline uuid.UUID) error {
	if err := q.LockRatings(ctx, discipline); err != nil {
		return err
	}
	matches, err := ratedMatches(ctx, q, discipline)
	if err != nil {
		return err
	}
	values := map[uuid.UUID]ratingValue{}
	for _, m := range matches {
		home, away := values[m.home], values[m.away]
		if home.played == 0 {
			home.value = model.InitialRating
		}
		if away.played == 0 {
			away.value = model.InitialRating
		}
		score := 0.5
		if m.homeScore > m.awayScore {
			score = 1
		} else if m.homeScore < m.awayScore {
			score = 0
		}
		before := model.EloRating{Value: home.value}
		change, err := before.AfterMatch(away.value, score)
		if err != nil {
			return err
		}
		if err = saveRatingChange(ctx, q, discipline, m, m.home, home.value, change.Delta, home.played+1); err != nil {
			return err
		}
		if err = saveRatingChange(ctx, q, discipline, m, m.away, away.value, -change.Delta, away.played+1); err != nil {
			return err
		}
		values[m.home] = ratingValue{home.value + change.Delta, home.played + 1}
		values[m.away] = ratingValue{away.value - change.Delta, away.played + 1}
	}
	return nil
}
func ratedMatches(ctx context.Context, q repository.Store, discipline uuid.UUID) ([]ratingMatch, error) {
	records, err := q.ReadRatedMatches(ctx, discipline)
	if err != nil {
		return nil, err
	}
	matches := make([]ratingMatch, len(records))
	for i, r := range records {
		matches[i] = ratingMatch{r.ID, r.TournamentID, r.HomeTeamID, r.AwayTeamID, r.HomeScore, r.AwayScore}
	}
	return matches, nil
}
func saveRatingChange(ctx context.Context, q repository.Store, discipline uuid.UUID, m ratingMatch, team uuid.UUID, before, delta float64, played int) error {
	if err := q.UpsertRating(ctx, team, discipline, before+delta, played); err != nil {
		return err
	}
	previous, err := q.ReadLatestRatingRevision(ctx, m.id, team)
	if err != nil && !ignoreNoRows(err) {
		return err
	}
	if err == nil && math.Abs(before-previous.Before) < 1e-9 && math.Abs(delta-previous.Delta) < 1e-9 {
		return nil
	}
	record := repository.RatedMatch{ID: m.id, TournamentID: m.tournament, HomeTeamID: m.home, AwayTeamID: m.away, HomeScore: m.homeScore, AwayScore: m.awayScore}
	return q.InsertRatingRevision(ctx, discipline, record, team, before, delta, previous.Revision+1)
}
func recordMatchRating(ctx context.Context, q repository.Store, tournament, discipline uuid.UUID, m ratingMatch, correction bool) error {
	if correction {
		return rebuildRatings(ctx, q, discipline)
	}
	if err := q.LockRatings(ctx, discipline); err != nil {
		return err
	}
	var err error
	m.home, m.away, err = q.MatchTeams(ctx, m.id)
	if err != nil {
		return err
	}
	m.tournament = tournament
	home, err := currentRating(ctx, q, m.home, discipline)
	if err != nil {
		return err
	}
	away, err := currentRating(ctx, q, m.away, discipline)
	if err != nil {
		return err
	}
	score := 0.5
	if m.homeScore > m.awayScore {
		score = 1
	} else if m.homeScore < m.awayScore {
		score = 0
	}
	rating := model.EloRating{Value: home.value}
	change, err := rating.AfterMatch(away.value, score)
	if err != nil {
		return err
	}
	if err = saveRatingChange(ctx, q, discipline, m, m.home, home.value, change.Delta, home.played+1); err != nil {
		return err
	}
	return saveRatingChange(ctx, q, discipline, m, m.away, away.value, -change.Delta, away.played+1)
}
func currentRating(ctx context.Context, q repository.Store, team, discipline uuid.UUID) (ratingValue, error) {
	snapshot, err := q.ReadRating(ctx, team, discipline)
	if ignoreNoRows(err) {
		return ratingValue{value: model.InitialRating}, nil
	}
	return ratingValue{snapshot.Value, snapshot.Played}, err
}
