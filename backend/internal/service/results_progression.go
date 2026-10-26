package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/apperror"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func updateProgression(ctx context.Context, q repository.Store, m model.Match, winner *uuid.UUID) error {
	if m.NextMatchId != nil {
		if err := advanceSlot(ctx, q, *m.NextMatchId, string(*m.NextMatchSlot), winner); err != nil {
			return err
		}
	}
	if m.LoserNextMatchId != nil {
		loser := m.HomeApplicationId
		if optionalUUIDEqual(loser, winner) {
			loser = m.AwayApplicationId
		}
		if err := advanceSlot(ctx, q, *m.LoserNextMatchId, string(*m.LoserNextMatchSlot), loser); err != nil {
			return err
		}
	}
	return nil
}
func advanceSlot(ctx context.Context, q repository.Store, id uuid.UUID, slot string, participant *uuid.UUID) error {
	if err := q.LockSchedule(ctx); err != nil {
		return err
	}
	next, err := q.ReadMatch(ctx, id, true)
	if err != nil {
		return err
	}
	existing := next.HomeApplicationId
	if slot == "away" {
		existing = next.AwayApplicationId
	}
	if optionalUUIDEqual(existing, participant) {
		return nil
	}
	if next.Status != model.MatchStatusScheduled || (existing != nil && next.StartsAt != nil && !next.StartsAt.After(time.Now())) {
		return conflict("winner cannot change after a dependent match has started")
	}
	if err = q.SetMatchParticipant(ctx, id, model.MatchNextMatchSlot(slot), participant); err != nil {
		return err
	}
	if slot == "away" {
		next.AwayApplicationId = participant
	} else {
		next.HomeApplicationId = participant
	}
	next.Version++
	return revalidateAdvancedSchedule(ctx, q, next)
}
func revalidateAdvancedSchedule(ctx context.Context, q repository.Store, next model.Match) error {
	if next.StartsAt == nil {
		return nil
	}
	t, err := q.ReadTournament(ctx, next.TournamentId, false)
	if err != nil {
		return err
	}
	err = scheduleConflict(ctx, q, t, next, *next.StartsAt)
	var domainErr *apperror.Error
	if err != nil && (!errors.As(err, &domainErr) || domainErr.Kind != apperror.Conflict) {
		return err
	}
	if err == nil && next.StartsAt.After(time.Now()) {
		return nil
	}
	if err = q.SetMatchSchedule(ctx, next.Id, nil, next.Venue, next.OnlineUrl); err != nil {
		return err
	}
	next.StartsAt = nil
	next.Version++
	if err = matchChanged(ctx, q, uuid.Nil, next, "match.unscheduled_after_advancement"); err != nil {
		return err
	}
	return notifyMatch(ctx, q, next, "match_unscheduled", "После определения участников требуется новое время матча")
}

func finishStage(ctx context.Context, q repository.Store, t model.Tournament, stageID uuid.UUID) error {
	incomplete, err := q.StageRemainingMatches(ctx, stageID)
	if err != nil {
		return err
	}
	if incomplete > 0 {
		return nil
	}
	kind, err := q.CompleteStage(ctx, stageID)
	if err != nil {
		return err
	}
	if kind == model.StageKindGroups {
		return startPlayoffs(ctx, q, t, stageID)
	}
	pending, err := q.RemainingTournamentStages(ctx, t.Id)
	if err != nil {
		return err
	}
	if pending == 0 {
		return finalizeTournament(ctx, q, t.Id)
	}
	return nil
}

func qualifiedParticipants(ctx context.Context, q repository.Store, t model.Tournament, stageID uuid.UUID) ([]uuid.UUID, error) {
	tables, err := q.ReadStandings(ctx, t.Id)
	if err != nil {
		return nil, err
	}
	relevant := stageStandings(tables, stageID)
	settings := t.Rules.GroupSettings
	selected := automaticQualifiers(relevant, settings.QualifiersPerGroup)
	if settings.WildcardCount == 0 {
		return selected, nil
	}
	wildcards := rankedGroupRows(relevant, *settings.WildcardPlace)
	sort.Slice(wildcards, func(i, j int) bool { return wildcardLess(wildcards[i], wildcards[j]) })
	for i := 0; i < settings.WildcardCount; i++ {
		selected = append(selected, wildcards[i].ApplicationId)
	}
	return selected, nil
}

func stageStandings(tables []model.Standings, stageID uuid.UUID) []model.Standings {
	result := []model.Standings{}
	for _, table := range tables {
		if table.StageId == stageID {
			result = append(result, table)
		}
	}
	return result
}

func rankedGroupRows(tables []model.Standings, rank int) []model.Standing {
	result := []model.Standing{}
	for _, table := range tables {
		for _, row := range table.Rows {
			if row.Position == rank {
				result = append(result, row)
			}
		}
	}
	return result
}

func automaticQualifiers(tables []model.Standings, count int) []uuid.UUID {
	selected := []uuid.UUID{}
	for rank := 1; rank <= count; rank++ {
		for _, row := range rankedGroupRows(tables, rank) {
			selected = append(selected, row.ApplicationId)
		}
	}
	return selected
}

func wildcardLess(a, b model.Standing) bool {
	if a.Points != b.Points {
		return a.Points > b.Points
	}
	if a.ScoreDifference != b.ScoreDifference {
		return a.ScoreDifference > b.ScoreDifference
	}
	if a.ScoreFor != b.ScoreFor {
		return a.ScoreFor > b.ScoreFor
	}
	return a.ApplicationId.String() < b.ApplicationId.String()
}

func startPlayoffs(ctx context.Context, tx repository.Store, t model.Tournament, groupStage uuid.UUID) error {
	ids, err := qualifiedParticipants(ctx, tx, t, groupStage)
	if err != nil {
		return err
	}
	stage, existing, err := readPlayoffStage(ctx, tx, t.Id)
	if err != nil {
		return err
	}
	matches, err := playoffMatches(t, stage, ids)
	if err != nil {
		return err
	}
	if err = tx.LockSchedule(ctx); err != nil {
		return err
	}
	return replacePlayoffDraw(ctx, tx, stage, existing, matches)
}

func readPlayoffStage(ctx context.Context, q repository.Store, tournamentID uuid.UUID) (model.Stage, []model.Match, error) {
	stage, err := q.ReadPlayoffStage(ctx, tournamentID)
	if err != nil {
		return stage, nil, err
	}
	matches, err := q.StageMatches(ctx, stage.Id)
	return stage, matches, err
}

func replacePlayoffDraw(ctx context.Context, q repository.Store, stage model.Stage, existing, matches []model.Match) error {
	if len(existing) > 0 {
		if firstRoundSame(existing, matches) {
			return nil
		}
		if err := requireUnstartedPlayoffs(existing); err != nil {
			return err
		}
		if err := q.DeleteStageMatches(ctx, stage.Id); err != nil {
			return err
		}
	}
	return persistPlayoffDraw(ctx, q, stage, matches)
}

func requireUnstartedPlayoffs(matches []model.Match) error {
	for _, m := range matches {
		if m.Status == model.MatchStatusCompleted || m.Status == model.MatchStatusAwaitingConfirmation || m.Status == model.MatchStatusDisputed || (m.StartsAt != nil && !m.StartsAt.After(time.Now())) {
			return conflict("group correction changes qualifiers after playoffs started")
		}
	}
	return nil
}

func persistPlayoffDraw(ctx context.Context, q repository.Store, stage model.Stage, matches []model.Match) error {
	if err := q.InsertMatches(ctx, matches); err != nil {
		return err
	}
	if err := q.SetStageStatus(ctx, stage.Id, model.StageStatusInProgress); err != nil {
		return err
	}
	return q.Event(ctx, stage.TournamentId, "bracket_updated", map[string]any{"stage_id": stage.Id, "matches": matches})
}

// finalizeTournament сохраняет места после завершения всех этапов и матчей.
func finalizeTournament(ctx context.Context, q repository.Store, id uuid.UUID) error {
	total, unfinished, err := q.TournamentMatchCounts(ctx, id)
	if err != nil {
		return err
	}
	if total == 0 || unfinished > 0 {
		return conflict("all matches must be completed")
	}
	tournament, err := q.ReadTournament(ctx, id, false)
	if err != nil {
		return err
	}
	if err = assignFinalPlaces(ctx, q, tournament); err != nil {
		return err
	}
	if err = q.CompleteTournamentStages(ctx, id); err != nil {
		return err
	}
	if err = q.SetTournamentStatus(ctx, id, model.TournamentStatusCompleted); err != nil {
		return err
	}
	if err = q.BumpTournamentVersion(ctx, id); err != nil {
		return err
	}
	return q.Event(ctx, id, "tournament_updated", map[string]any{"status": "completed"})
}

func assignFinalPlaces(ctx context.Context, q repository.Store, t model.Tournament) error {
	if t.Format == model.TournamentFormatRoundRobin {
		tables, err := q.ReadStandings(ctx, t.Id)
		if err != nil {
			return err
		}
		for _, table := range tables {
			for _, row := range table.Rows {
				if err = q.SetFinalPlace(ctx, row.ApplicationId, row.Position, false); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := q.ClearFinalPlaces(ctx, t.Id); err != nil {
		return err
	}
	if err := playoffPlaces(ctx, q, t); err != nil {
		return err
	}
	return unqualifiedPlaces(ctx, q, t.Id)
}
func unqualifiedPlaces(ctx context.Context, q repository.Store, tournamentID uuid.UUID) error {
	rows, err := q.UnplacedStandings(ctx, tournamentID)
	if err != nil {
		return err
	}
	base, err := q.PlacedApplicationsCount(ctx, tournamentID)
	if err != nil {
		return err
	}
	previous, rank := 0, 0
	for i, row := range rows {
		if row.Position != previous {
			rank = i + 1
			previous = row.Position
		}
		if err = q.SetFinalPlace(ctx, row.ApplicationId, base+rank, false); err != nil {
			return err
		}
	}
	return nil
}

func playoffPlaces(ctx context.Context, q repository.Store, t model.Tournament) error {
	matches, err := q.PlayoffMatches(ctx, t.Id)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return conflict("playoffs have no matches")
	}
	maxRound := matches[0].Round
	for _, m := range matches {
		if m.Status == model.MatchStatusBye || m.WinnerApplicationId == nil {
			continue
		}
		loser := m.HomeApplicationId
		if optionalUUIDEqual(loser, m.WinnerApplicationId) {
			loser = m.AwayApplicationId
		}
		place := 1 << (maxRound - m.Round)
		place++
		if m.Round == maxRound && m.Position == 1 {
			if err = q.SetFinalPlace(ctx, *m.WinnerApplicationId, 1, false); err != nil {
				return err
			}
			place = 2
		}
		if m.Round == maxRound && m.Position == 2 {
			if err = q.SetFinalPlace(ctx, *m.WinnerApplicationId, 3, false); err != nil {
				return err
			}
			place = 4
		}
		if loser != nil {
			if err = q.SetFinalPlace(ctx, *loser, place, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func firstRoundSame(existing, expected []model.Match) bool {
	old := []model.Match{}
	want := []model.Match{}
	for _, m := range existing {
		if m.Round == 1 {
			old = append(old, m)
		}
	}
	for _, m := range expected {
		if m.Round == 1 {
			want = append(want, m)
		}
	}
	if len(old) != len(want) {
		return false
	}
	for i := range old {
		if old[i].Position != want[i].Position || !optionalUUIDEqual(old[i].HomeApplicationId, want[i].HomeApplicationId) || !optionalUUIDEqual(old[i].AwayApplicationId, want[i].AwayApplicationId) {
			return false
		}
	}
	return true
}
