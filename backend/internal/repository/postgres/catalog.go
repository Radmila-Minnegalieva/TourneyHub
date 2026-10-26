package postgres

import (
	"context"
	"strings"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/google/uuid"
)

func (r *Repository) DisciplinePage(ctx context.Context, limit, offset int) (model.DisciplinePage, error) {
	page := model.DisciplinePage{Limit: limit, Offset: offset}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM disciplines WHERE active`).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	items, err := jsonItems[model.Discipline](ctx, r.db, `SELECT to_jsonb(d) FROM disciplines d WHERE active ORDER BY lower(name),id LIMIT $1 OFFSET $2`, limit, offset)
	page.Items = items
	return page, err
}

func (r *Repository) CreateDisciplineRecord(ctx context.Context, request model.DisciplineRequest) (model.Discipline, error) {
	return jsonObject[model.Discipline](ctx, r.db, `INSERT INTO disciplines(name,min_roster_size,max_roster_size,allows_draw,score_unit) VALUES($1,$2,$3,$4,$5) RETURNING to_jsonb(disciplines)`, strings.TrimSpace(request.Name), request.MinRosterSize, request.MaxRosterSize, request.AllowsDraw, strings.TrimSpace(request.ScoreUnit))
}

func (r *Repository) Discipline(ctx context.Context, id uuid.UUID, lock bool) (model.Discipline, error) {
	query := `SELECT to_jsonb(d) FROM disciplines d WHERE id=$1`
	if lock {
		query += ` FOR UPDATE`
	}
	return jsonObject[model.Discipline](ctx, r.db, query, id)
}

func (r *Repository) UpdateDisciplineRecord(ctx context.Context, id uuid.UUID, discipline model.Discipline) (model.Discipline, error) {
	return jsonObject[model.Discipline](ctx, r.db, `UPDATE disciplines SET name=$2,min_roster_size=$3,max_roster_size=$4,allows_draw=$5,score_unit=$6,active=$7,version=version+1 WHERE id=$1 RETURNING to_jsonb(disciplines)`, id, discipline.Name, discipline.MinRosterSize, discipline.MaxRosterSize, discipline.AllowsDraw, discipline.ScoreUnit, discipline.Active)
}

func (r *Repository) RatingPage(ctx context.Context, discipline uuid.UUID, limit, offset int) (model.RatingPage, error) {
	page := model.RatingPage{Limit: limit, Offset: offset}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM ratings r JOIN teams t ON t.id=r.team_id WHERE r.discipline_id=$1 AND t.deleted_at IS NULL`, discipline).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	items, err := jsonItems[model.Rating](ctx, r.db, `SELECT jsonb_build_object('discipline_id',r.discipline_id,'played',r.played,'position',row_number() OVER(ORDER BY r.value DESC,t.id),'value',r.value,'team',to_jsonb(t)||jsonb_build_object('logo_url',CASE WHEN t.logo_file_id IS NOT NULL THEN '/api/v1/files/'||t.logo_file_id END)) FROM ratings r JOIN teams t ON t.id=r.team_id WHERE r.discipline_id=$1 AND t.deleted_at IS NULL ORDER BY r.value DESC,t.id LIMIT $2 OFFSET $3`, discipline, limit, offset)
	page.Items = items
	return page, err
}

func (r *Repository) RatingHistoryPage(ctx context.Context, team, discipline uuid.UUID, limit, offset int) (model.RatingChangePage, error) {
	page := model.RatingChangePage{Limit: limit, Offset: offset}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM rating_changes WHERE team_id=$1 AND discipline_id=$2`, team, discipline).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	items, err := jsonItems[model.RatingChange](ctx, r.db, `SELECT to_jsonb(r) FROM rating_changes r WHERE team_id=$1 AND discipline_id=$2 ORDER BY created_at DESC,id LIMIT $3 OFFSET $4`, team, discipline, limit, offset)
	page.Items = items
	return page, err
}

func (r *Repository) NotificationPage(ctx context.Context, user uuid.UUID, unread *bool, limit, offset int) (model.NotificationPage, error) {
	page := model.NotificationPage{Limit: limit, Offset: offset}
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM notifications WHERE user_id=$1 AND ($2::boolean IS NOT TRUE OR read_at IS NULL)`, user, unread).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	items, err := jsonItems[model.Notification](ctx, r.db, `SELECT to_jsonb(n) FROM notifications n WHERE user_id=$1 AND ($2::boolean IS NOT TRUE OR read_at IS NULL) ORDER BY created_at DESC,id LIMIT $3 OFFSET $4`, user, unread, limit, offset)
	page.Items = items
	return page, err
}

func (r *Repository) MarkNotificationRead(ctx context.Context, id, user uuid.UUID) (bool, error) {
	tag, err := r.db.Exec(ctx, `UPDATE notifications SET read_at=coalesce(read_at,now()) WHERE id=$1 AND user_id=$2`, id, user)
	return tag.RowsAffected() > 0, dbError(err)
}
