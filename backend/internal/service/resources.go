package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/csvutil"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/fileutil"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/security"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"

	"github.com/google/uuid"
)

const MaxFileBytes = 5 * 1024 * 1024

func (s *Service) UploadFile(ctx context.Context, upload model.FileUpload) (model.File, error) {
	var result model.File
	if err := requireRole(ctx, model.Participant); err != nil {
		return result, err
	}
	if !upload.Purpose.Valid() {
		return result, invalid("invalid file purpose")
	}
	contentType, err := fileutil.Validate(upload.Data, upload.Purpose == model.FilePurposeMatchProtocol, MaxFileBytes)
	if err != nil {
		return result, err
	}
	upload.Filename, err = fileutil.CleanName(upload.Filename)
	if err != nil {
		return result, err
	}
	err = s.withTx(ctx, func(q repository.Store) error {
		var err error
		result, err = q.CreateFile(ctx, uuid.New(), currentUser(ctx), upload, contentType)
		return err
	})
	return result, err
}
func (s *Service) DownloadFile(ctx context.Context, id uuid.UUID) (model.FileContent, error) {
	file, err := s.repo.File(ctx, id)
	if err != nil {
		return model.FileContent{}, err
	}
	if file.File.Purpose == model.FilePurposeMatchProtocol && file.OwnerID != currentUser(ctx) && !hasRole(ctx, model.Admin) {
		if currentUser(ctx) == uuid.Nil {
			return model.FileContent{}, forbidden("file is private")
		}
		allowed, err := s.repo.CanReadProtocol(ctx, id, currentUser(ctx))
		if err != nil {
			return model.FileContent{}, err
		}
		if !allowed {
			return model.FileContent{}, forbidden("file is private")
		}
	}
	return model.FileContent{File: file.File, Data: file.Data}, nil
}
func (s *Service) ExportTournament(ctx context.Context, id uuid.UUID) ([]byte, error) {
	if err := tournamentPublic(ctx, s.repo, id); err != nil {
		return nil, err
	}
	tournament, err := s.repo.ReadTournament(ctx, id, false)
	if err != nil {
		return nil, err
	}
	if tournament.Status != model.TournamentStatusCompleted {
		return nil, conflict("tournament is not completed")
	}
	rows, err := s.repo.ExportRows(ctx, id)
	if err != nil {
		return nil, err
	}
	var buffer bytes.Buffer
	buffer.WriteString("\xef\xbb\xbf")
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"row_type", "place", "team", "match_id", "home_team", "away_team", "home_score", "away_score", "status"}); err != nil {
		return nil, err
	}
	for _, row := range rows {
		record := []string{row.Kind, row.Place, row.Team, row.MatchID, row.HomeTeam, row.AwayTeam, row.HomeScore, row.AwayScore, row.Status}
		for i := range record {
			record[i] = csvutil.EscapeCell(record[i])
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	return buffer.Bytes(), writer.Error()
}
func (s *Service) TournamentVisible(ctx context.Context, id uuid.UUID) error {
	return tournamentPublic(ctx, s.repo, id)
}
func (s *Service) LatestEventID(ctx context.Context, id uuid.UUID) (int64, error) {
	_, max, err := s.repo.EventBounds(ctx, id)
	return max, err
}
func (s *Service) ReplayEvents(ctx context.Context, id uuid.UUID, after int64) (model.EventBatch, error) {
	min, max, err := s.repo.EventBounds(ctx, id)
	if err != nil {
		return model.EventBatch{}, err
	}
	if after > max || (after > 0 && min > after) {
		return model.EventBatch{Cursor: max, Resync: true, Events: []model.TournamentEvent{}}, nil
	}
	events, err := s.repo.EventsAfter(ctx, id, after, 100)
	if err != nil {
		return model.EventBatch{}, err
	}
	if len(events) > 0 {
		after = events[len(events)-1].ID
	}
	return model.EventBatch{Events: events, Cursor: after}, nil
}
func (s *Service) BootstrapAdmin(ctx context.Context, email, nickname, password string) error {
	hash, err := security.HashPassword(password)
	if err != nil {
		return err
	}
	return s.withTx(ctx, func(q repository.Store) error {
		if err := q.LockAdmin(ctx); err != nil {
			return err
		}
		state, err := q.BootstrapState(ctx, email)
		if err != nil {
			return err
		}
		if state.AdminCount > 0 {
			if state.SameActiveAdmin {
				return nil
			}
			return conflict("an administrator already exists; use the authenticated admin API")
		}
		id := uuid.New()
		if err := q.CreateBootstrapAdmin(ctx, id, email, nickname, hash); err != nil {
			return err
		}
		return q.Audit(ctx, id, "administrator_bootstrapped", "user", id, nil, map[string]any{"bootstrap": true})
	})
}

func (s *Service) Cleanup(ctx context.Context) error {
	if err := s.repo.CleanupRateLimits(ctx); err != nil {
		return err
	}
	return s.repo.CleanupEvents(ctx, time.Now().Add(-s.cfg.Worker.EventRetention))
}
