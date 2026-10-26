package api

import (
	"fmt"
	contract "github.com/Radmila-Minnegalieva/TourneyHub/backend/pkg/api/v1"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
	"time"
)

func (a *API) ExportTournament(c echo.Context, id uuid.UUID) error {
	data, err := a.resources.ExportTournament(c.Request().Context(), id)
	if err != nil {
		return err
	}
	c.Response().Header().Set("Content-Disposition", `attachment; filename="tournament-`+id.String()+`.csv"`)
	return c.Blob(200, "text/csv; charset=utf-8", data)
}
func (a *API) StreamTournamentEvents(c echo.Context, id uuid.UUID, p contract.StreamTournamentEventsParams) error {
	ctx := c.Request().Context()
	if err := a.resources.TournamentVisible(ctx, id); err != nil {
		return err
	}
	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return echo.NewHTTPError(500, "streaming unavailable")
	}
	after, err := parseEventCursor(p.LastEventID)
	if err != nil {
		return err
	}
	if p.LastEventID == nil {
		after, err = a.resources.LatestEventID(ctx, id)
		if err != nil {
			return err
		}
	}
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("X-Accel-Buffering", "no")
	c.Response().WriteHeader(200)
	if _, err := fmt.Fprint(c.Response(), ": connected\n\n"); err != nil {
		return nil
	}
	flusher.Flush()
	return a.serveEvents(c, id, after, flusher)
}
func parseEventCursor(value *string) (int64, error) {
	if value == nil {
		return 0, nil
	}
	after, err := strconv.ParseInt(*value, 10, 64)
	if err != nil || after < 0 {
		return 0, echo.NewHTTPError(400, "invalid Last-Event-ID")
	}
	return after, nil
}
func (a *API) serveEvents(c echo.Context, id uuid.UUID, after int64, flusher http.Flusher) error {
	ctx := c.Request().Context()
	tick := time.NewTicker(a.cfg.SSEPollInterval)
	defer tick.Stop()
	heartbeat := time.NewTicker(a.cfg.SSEHeartbeatInterval)
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-heartbeat.C:
			if _, err := fmt.Fprint(c.Response(), ": heartbeat\n\n"); err != nil {
				return nil
			}
			flusher.Flush()
		case <-tick.C:
			if err := a.resources.TournamentVisible(ctx, id); err != nil {
				return nil
			}
			next, err := a.replayEvents(c, id, after)
			if err != nil {
				return err
			}
			after = next
			flusher.Flush()
		}
	}
}
func (a *API) replayEvents(c echo.Context, id uuid.UUID, after int64) (int64, error) {
	batch, err := a.resources.ReplayEvents(c.Request().Context(), id, after)
	if err != nil {
		return after, err
	}
	if batch.Resync {
		_, err := fmt.Fprintf(c.Response(), "id: %d\nevent: resync_required\ndata: {}\n\n", batch.Cursor)
		return batch.Cursor, err
	}
	for _, event := range batch.Events {
		if _, err := fmt.Fprintf(c.Response(), "id: %d\nevent: %s\ndata: %s\n\n", event.ID, event.Kind, event.Payload); err != nil {
			return after, err
		}
		after = event.ID
	}
	return after, nil
}
