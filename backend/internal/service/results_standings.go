package service

import (
	"context"
	"sort"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

type tableRow struct {
	model.Standing
	group *uuid.UUID
	elo   float64
	mini  int
}

func standingsParticipants(ctx context.Context, q repository.Store, t model.Tournament, stageID uuid.UUID) ([]*tableRow, error) {
	participants, err := q.ReadStandingParticipants(ctx, t.Id, stageID, t.DisciplineId)
	if err != nil {
		return nil, err
	}
	result := make([]*tableRow, 0, len(participants))
	for _, p := range participants {
		result = append(result, &tableRow{Standing: model.Standing{ApplicationId: p.ApplicationID}, group: p.GroupID, elo: p.Elo})
	}
	return result, nil
}

func updateStanding(r *tableRow, home, away int, rules model.Rules) {
	r.Played++
	r.ScoreFor += home
	r.ScoreAgainst += away
	r.ScoreDifference = r.ScoreFor - r.ScoreAgainst
	switch {
	case home > away:
		r.Wins++
		r.Points += rules.WinPoints
	case home < away:
		r.Losses++
		r.Points += rules.LossPoints
	default:
		r.Draws++
		r.Points += rules.DrawPoints
	}
}
func headToHead(rows map[uuid.UUID]*tableRow, matches []model.Match, rules model.Rules) {
	for _, m := range matches {
		home, away := rows[*m.HomeApplicationId], rows[*m.AwayApplicationId]
		if home == nil || away == nil || home.Points != away.Points {
			continue
		}
		switch {
		case *m.HomeScore > *m.AwayScore:
			home.mini += rules.WinPoints
			away.mini += rules.LossPoints
		case *m.HomeScore < *m.AwayScore:
			away.mini += rules.WinPoints
			home.mini += rules.LossPoints
		default:
			home.mini += rules.DrawPoints
			away.mini += rules.DrawPoints
		}
	}
}
func standingLess(a, b *tableRow, breakers []model.TieBreaker) bool {
	if a.Points != b.Points {
		return a.Points > b.Points
	}
	for _, breaker := range breakers {
		av, bv := 0.0, 0.0
		switch breaker {
		case model.TieBreakerHeadToHead:
			av, bv = float64(a.mini), float64(b.mini)
		case model.TieBreakerScoreDifference:
			av, bv = float64(a.ScoreDifference), float64(b.ScoreDifference)
		case model.TieBreakerScoreFor:
			av, bv = float64(a.ScoreFor), float64(b.ScoreFor)
		case model.TieBreakerElo:
			av, bv = a.elo, b.elo
		case model.TieBreakerDraw:
			return a.ApplicationId.String() < b.ApplicationId.String()
		}
		if av != bv {
			return av > bv
		}
	}
	return a.ApplicationId.String() < b.ApplicationId.String()
}
func rebuildStandings(ctx context.Context, q repository.Store, t model.Tournament, stageID uuid.UUID) error {
	participants, err := standingsParticipants(ctx, q, t, stageID)
	if err != nil {
		return err
	}
	if len(participants) == 0 {
		return nil
	}
	rowsByID := map[uuid.UUID]*tableRow{}
	buckets := map[uuid.UUID][]*tableRow{}
	for _, row := range participants {
		rowsByID[row.ApplicationId] = row
		key := uuid.Nil
		if row.group != nil {
			key = *row.group
		}
		buckets[key] = append(buckets[key], row)
	}
	matches, err := q.CompletedStageMatches(ctx, stageID)
	if err != nil {
		return err
	}
	for _, m := range matches {
		home, away := rowsByID[*m.HomeApplicationId], rowsByID[*m.AwayApplicationId]
		if home == nil || away == nil {
			continue
		}
		updateStanding(home, *m.HomeScore, *m.AwayScore, t.Rules)
		updateStanding(away, *m.AwayScore, *m.HomeScore, t.Rules)
	}
	headToHead(rowsByID, matches, t.Rules)
	if err = q.ClearStageStandings(ctx, stageID); err != nil {
		return err
	}
	for _, bucket := range buckets {
		sort.Slice(bucket, func(i, j int) bool { return standingLess(bucket[i], bucket[j], t.Rules.TieBreakers) })
		for i, row := range bucket {
			row.Position = i + 1
			if err = writeStanding(ctx, q, t.Id, stageID, row); err != nil {
				return err
			}
		}
	}
	return q.Event(ctx, t.Id, "standings_updated", map[string]any{"stage_id": stageID})
}
func writeStanding(ctx context.Context, q repository.Store, tournamentID, stageID uuid.UUID, row *tableRow) error {
	return q.InsertStanding(ctx, tournamentID, stageID, row.group, row.Standing)
}

func (s *Service) GetStandings(ctx context.Context, id uuid.UUID) ([]model.Standings, error) {
	if err := tournamentPublic(ctx, s.repo, id); err != nil {
		return nil, err
	}
	return s.repo.ReadStandings(ctx, id)
}
