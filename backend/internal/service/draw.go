package service

import (
	"context"
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

func acceptedIDs(ctx context.Context, q repository.Store, t model.Tournament) ([]uuid.UUID, error) {
	return q.AcceptedApplicationsByRating(ctx, t.Id, t.DisciplineId)
}

func orderSeeds(ids []uuid.UUID, req model.SeedingRequest) ([]uuid.UUID, error) {
	if req.Method == model.SeedingRequestMethodManual {
		if req.ApplicationIds == nil || len(*req.ApplicationIds) != len(ids) {
			return nil, invalid("manual seeding needs all accepted applications")
		}
		expected := map[uuid.UUID]bool{}
		for _, id := range ids {
			expected[id] = true
		}
		for _, id := range *req.ApplicationIds {
			if !expected[id] {
				return nil, invalid("manual seeding must be a permutation")
			}
			delete(expected, id)
		}
		return *req.ApplicationIds, nil
	}
	if req.Method != model.SeedingRequestMethodRandom {
		return ids, nil
	}
	for i := len(ids) - 1; i > 0; i-- {
		j, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return nil, err
		}
		ids[i], ids[j.Int64()] = ids[j.Int64()], ids[i]
	}
	return ids, nil
}
func drawAvailable(ctx context.Context, q repository.Store, t model.Tournament) error {
	if t.Status != model.TournamentStatusRegistration && t.Status != model.TournamentStatusInProgress {
		return conflict("draw requires a published active tournament")
	}
	started, err := q.HasTournamentReports(ctx, t.Id)
	if err != nil {
		return err
	}
	if started {
		return conflict("cannot rebuild after the first result")
	}
	return nil
}

func buildDraw(t model.Tournament, ids []uuid.UUID) (model.DrawPreview, error) {
	p := model.DrawPreview{Id: uuid.New(), TournamentId: t.Id, TournamentVersion: t.Version, ExpiresAt: time.Now().UTC().Add(time.Hour), Stages: []model.Stage{}, Groups: []model.Group{}, Matches: []model.Match{}, Seeds: []model.Seed{}}
	for i, id := range ids {
		p.Seeds = append(p.Seeds, model.Seed{ApplicationId: id, Position: i + 1})
	}
	switch t.Format {
	case model.TournamentFormatSingleElimination:
		if t.Rules.ThirdPlaceMatch && len(ids) < 4 {
			return p, invalid("third-place match needs at least four participants")
		}
		stage := newStage(t.Id, model.StageKindPlayoffs, 1, model.StageStatusInProgress)
		p.Stages = append(p.Stages, stage)
		matches, err := playoffMatches(t, stage, ids)
		p.Matches = matches
		return p, err
	case model.TournamentFormatRoundRobin:
		stage := newStage(t.Id, model.StageKindRoundRobin, 1, model.StageStatusInProgress)
		p.Stages = append(p.Stages, stage)
		matches, err := roundRobinMatches(t, stage, nil, ids)
		p.Matches = matches
		return p, err
	default:
		return buildGroupDraw(t, ids, p)
	}
}
func newStage(tid uuid.UUID, kind model.StageKind, pos int, status model.StageStatus) model.Stage {
	return model.Stage{Id: uuid.New(), TournamentId: tid, Kind: kind, Position: pos, Status: status}
}
func baseMatch(t model.Tournament, s model.Stage, round, pos int) model.Match {
	return model.Match{Id: uuid.New(), TournamentId: t.Id, StageId: s.Id, Round: round, Position: pos, Status: model.MatchStatusScheduled, Version: 1}
}
func roundRobinMatches(t model.Tournament, s model.Stage, gid *uuid.UUID, ids []uuid.UUID) ([]model.Match, error) {
	pairs, err := model.RoundRobin(len(ids), int(t.Rules.RoundRobinLegs))
	if err != nil {
		return nil, invalid(err.Error())
	}
	result := make([]model.Match, 0, len(pairs))
	for _, pair := range pairs {
		m := baseMatch(t, s, pair.Round, pair.Position)
		m.HomeApplicationId = &ids[pair.Home]
		m.AwayApplicationId = &ids[pair.Away]
		m.GroupId = gid
		result = append(result, m)
	}
	return result, nil
}
func playoffMatches(t model.Tournament, s model.Stage, ids []uuid.UUID) ([]model.Match, error) {
	order, err := model.SeedOrder(len(ids))
	if err != nil {
		return nil, invalid(err.Error())
	}
	rounds := [][]model.Match{}
	for count, round := len(order)/2, 1; count >= 1; count, round = count/2, round+1 {
		matches := make([]model.Match, count)
		for i := range matches {
			matches[i] = baseMatch(t, s, round, i+1)
		}
		rounds = append(rounds, matches)
	}
	for i := range rounds[0] {
		m := &rounds[0][i]
		h, a := order[2*i]-1, order[2*i+1]-1
		if h < len(ids) {
			m.HomeApplicationId = &ids[h]
		}
		if a < len(ids) {
			m.AwayApplicationId = &ids[a]
		}
	}
	linkPlayoffRounds(rounds)
	result := []model.Match{}
	for _, round := range rounds {
		result = append(result, round...)
	}
	if t.Rules.ThirdPlaceMatch && len(rounds) > 1 {
		third := baseMatch(t, s, len(rounds), 2)
		semis := len(result) - 3
		home, away := model.MatchLoserNextMatchSlot("home"), model.MatchLoserNextMatchSlot("away")
		result[semis].LoserNextMatchId = &third.Id
		result[semis].LoserNextMatchSlot = &home
		result[semis+1].LoserNextMatchId = &third.Id
		result[semis+1].LoserNextMatchSlot = &away
		result = append(result, third)
	}
	for i := 0; i < len(order)/2; i++ {
		m := &result[i]
		if m.HomeApplicationId == nil || m.AwayApplicationId == nil {
			m.Status = model.MatchStatusBye
			m.WinnerApplicationId = m.HomeApplicationId
			if m.WinnerApplicationId == nil {
				m.WinnerApplicationId = m.AwayApplicationId
			}
			fillWinner(result, *m)
		}
	}
	return result, nil
}
func linkPlayoffRounds(rounds [][]model.Match) {
	for r := 0; r < len(rounds)-1; r++ {
		for i := range rounds[r] {
			next := &rounds[r+1][i/2]
			slot := model.MatchNextMatchSlot("home")
			if i%2 == 1 {
				slot = "away"
			}
			rounds[r][i].NextMatchId = &next.Id
			rounds[r][i].NextMatchSlot = &slot
		}
	}
}
func fillWinner(matches []model.Match, m model.Match) {
	if m.NextMatchId == nil {
		return
	}
	for i := range matches {
		if matches[i].Id == *m.NextMatchId {
			if *m.NextMatchSlot == "home" {
				matches[i].HomeApplicationId = m.WinnerApplicationId
			} else {
				matches[i].AwayApplicationId = m.WinnerApplicationId
			}
			return
		}
	}
}
func buildGroupDraw(t model.Tournament, ids []uuid.UUID, p model.DrawPreview) (model.DrawPreview, error) {
	gs := t.Rules.GroupSettings
	if gs == nil {
		return p, invalid("group settings are required")
	}
	groups, err := model.SnakeGroups(len(ids), gs.GroupCount)
	if err != nil {
		return p, invalid(err.Error())
	}
	if t.Rules.ThirdPlaceMatch && gs.GroupCount*gs.QualifiersPerGroup+gs.WildcardCount < 4 {
		return p, invalid("third-place match needs at least four playoff qualifiers")
	}
	minSize := len(ids) / gs.GroupCount
	if gs.QualifiersPerGroup > minSize || gs.QualifiersPerGroup < 1 {
		return p, invalid("qualifiers cannot exceed the number of participants in a group")
	}
	if gs.WildcardCount > 0 && (gs.WildcardPlace == nil || *gs.WildcardPlace <= gs.QualifiersPerGroup || *gs.WildcardPlace > minSize || gs.WildcardCount > gs.GroupCount) {
		return p, invalid("invalid wildcard settings")
	}
	stage := newStage(t.Id, model.StageKindGroups, 1, model.StageStatusInProgress)
	p.Stages = append(p.Stages, stage, newStage(t.Id, model.StageKindPlayoffs, 2, model.StageStatusPending))
	for i, indices := range groups {
		group := model.Group{Id: uuid.New(), StageId: stage.Id, Name: fmt.Sprintf("%c", 'A'+rune(i)), ApplicationIds: []uuid.UUID{}}
		for _, j := range indices {
			group.ApplicationIds = append(group.ApplicationIds, ids[j])
		}
		matches, e := roundRobinMatches(t, stage, &group.Id, group.ApplicationIds)
		if e != nil {
			return p, e
		}
		// position уникальна в пределах этапа и тура, включая разные группы.
		for j := range matches {
			matches[j].Position += i * 128
		}
		p.Groups = append(p.Groups, group)
		p.Matches = append(p.Matches, matches...)
	}
	return p, nil
}
func (s *Service) PreviewDraw(ctx context.Context, id uuid.UUID, req model.SeedingRequest) (model.DrawPreview, error) {
	var result model.DrawPreview
	err := s.withTx(ctx, func(tx repository.Store) error {
		tournament, err := tx.ReadTournament(ctx, id, true)
		if err != nil {
			return err
		}
		if err = requireTournamentOwner(ctx, tournament); err != nil {
			return err
		}
		if tournament.Version != req.Version {
			return conflict("tournament version is stale")
		}
		if err = drawAvailable(ctx, tx, tournament); err != nil {
			return err
		}
		ids, err := acceptedIDs(ctx, tx, tournament)
		if err != nil {
			return err
		}
		if len(ids) < 2 {
			return invalid("draw requires at least two accepted applications")
		}
		ids, err = orderSeeds(ids, req)
		if err != nil {
			return err
		}
		result, err = buildDraw(tournament, ids)
		if err != nil {
			return err
		}
		return tx.InsertDrawPreview(ctx, result, currentUser(ctx), req.Method)
	})
	return result, err
}

func (s *Service) ConfirmDraw(ctx context.Context, id uuid.UUID, req model.ConfirmDrawRequest) (model.Bracket, error) {
	var result model.Bracket
	err := s.withTx(ctx, func(tx repository.Store) error {
		var err error
		result, err = confirmPersistedDraw(ctx, tx, id, req)
		return err
	})
	return result, err
}

func confirmPersistedDraw(ctx context.Context, q repository.Store, id uuid.UUID, req model.ConfirmDrawRequest) (model.Bracket, error) {
	var result model.Bracket
	tournament, err := q.ReadTournament(ctx, id, true)
	if err != nil {
		return result, err
	}
	if err = requireTournamentOwner(ctx, tournament); err != nil {
		return result, err
	}
	preview, confirmed, err := q.ReadDrawPreview(ctx, id, req.PreviewId)
	if err != nil {
		return result, err
	}
	if confirmed {
		return q.ReadBracket(ctx, id)
	}
	if err = validatePersistedDraw(ctx, q, tournament, preview, req.Version); err != nil {
		return result, err
	}
	return materializeDraw(ctx, q, currentUser(ctx), preview, req.PreviewId)
}

func validatePersistedDraw(ctx context.Context, q repository.Store, t model.Tournament, p model.DrawPreview, version int) error {
	if t.Version != version || p.TournamentVersion != t.Version {
		return conflict("draw preview version is stale")
	}
	if time.Now().After(p.ExpiresAt) {
		return conflict("draw preview expired")
	}
	if err := drawAvailable(ctx, q, t); err != nil {
		return err
	}
	ids, err := acceptedIDs(ctx, q, t)
	if err != nil {
		return err
	}
	if !sameDrawParticipants(ids, p.Seeds) {
		return conflict("accepted participants changed")
	}
	return nil
}

func materializeDraw(ctx context.Context, q repository.Store, actor uuid.UUID, p model.DrawPreview, previewID uuid.UUID) (model.Bracket, error) {
	var result model.Bracket
	if err := q.ClearTournamentDraw(ctx, p.TournamentId); err != nil {
		return result, err
	}
	if err := saveDraw(ctx, q, p); err != nil {
		return result, err
	}
	if err := q.ConfirmDrawPreview(ctx, previewID); err != nil {
		return result, err
	}
	if err := q.SetTournamentStatus(ctx, p.TournamentId, model.TournamentStatusInProgress); err != nil {
		return result, err
	}
	if err := q.BumpTournamentVersion(ctx, p.TournamentId); err != nil {
		return result, err
	}
	if err := q.Audit(ctx, actor, "draw.confirmed", "tournament", p.TournamentId, nil, p); err != nil {
		return result, err
	}
	result = model.Bracket{Stages: p.Stages, Groups: p.Groups, Matches: p.Matches}
	return result, q.Event(ctx, p.TournamentId, "bracket_updated", result)
}

func sameDrawParticipants(ids []uuid.UUID, seeds []model.Seed) bool {
	if len(ids) != len(seeds) {
		return false
	}
	set := map[uuid.UUID]bool{}
	for _, id := range ids {
		set[id] = true
	}
	for _, s := range seeds {
		if !set[s.ApplicationId] {
			return false
		}
		delete(set, s.ApplicationId)
	}
	return len(set) == 0
}

func saveDraw(ctx context.Context, q repository.Store, p model.DrawPreview) error {
	for _, seed := range p.Seeds {
		if err := q.SetApplicationSeed(ctx, seed.ApplicationId, seed.Position); err != nil {
			return err
		}
	}
	for _, stage := range p.Stages {
		if err := q.InsertStage(ctx, stage); err != nil {
			return err
		}
	}
	for _, group := range p.Groups {
		if err := q.InsertGroup(ctx, p.TournamentId, group); err != nil {
			return err
		}
	}
	if err := q.InsertMatches(ctx, p.Matches); err != nil {
		return err
	}
	tournament, err := q.ReadTournament(ctx, p.TournamentId, false)
	if err != nil {
		return err
	}
	for _, stage := range p.Stages {
		if stage.Kind != model.StageKindPlayoffs {
			if err = rebuildStandings(ctx, q, tournament, stage.Id); err != nil {
				return err
			}
		}
	}
	return nil
}
