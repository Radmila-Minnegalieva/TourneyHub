package postgres

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/lib/pagination"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// tournamentJSON выдаёт только публичные поля организатора.
const tournamentJSON = `jsonb_build_object(
 'id',t.id,'organizer',jsonb_build_object('id',u.id,'nickname',u.nickname,'name',u.name,
 'avatar_url',CASE WHEN u.avatar_file_id IS NOT NULL THEN '/api/v1/files/'||u.avatar_file_id END),
 'name',t.name,'description',t.description,'discipline_id',t.discipline_id,'format',t.format,
 'max_participants',t.max_participants,'registration_opens_at',t.registration_opens_at,
 'registration_closes_at',t.registration_closes_at,'starts_at',t.starts_at,'ends_at',t.ends_at,
 'timezone',t.timezone,'status',t.status,'hidden',t.hidden,'referee_id',t.referee_id,
 'accepted_count',(SELECT count(*) FROM applications WHERE tournament_id=t.id AND status='accepted'),
 'version',t.version,'created_at',t.created_at,'rules',jsonb_build_object(
 'best_of',t.best_of,'round_robin_legs',t.round_robin_legs,'win_points',t.win_points,
 'draw_points',t.draw_points,'loss_points',t.loss_points,'tie_breakers',t.tie_breakers,
 'third_place_match',t.third_place_match,'confirmation_timeout_hours',t.confirmation_timeout_hours,
 'group_settings',CASE WHEN t.group_count IS NOT NULL THEN jsonb_build_object(
 'group_count',t.group_count,'qualifiers_per_group',t.qualifiers_per_group,
 'wildcard_count',t.wildcard_count,'wildcard_place',t.wildcard_place) END))`

func (r *Repository) ReadTournament(ctx context.Context, id uuid.UUID, lock bool) (model.Tournament, error) {
	query := `SELECT ` + tournamentJSON + ` FROM tournaments t JOIN users u ON u.id=t.organizer_id WHERE t.id=$1 AND t.deleted_at IS NULL`
	if lock {
		query += ` FOR UPDATE OF t`
	}
	var raw []byte
	var tournament model.Tournament
	if err := r.db.QueryRow(ctx, query, id).Scan(&raw); err != nil {
		return tournament, dbError(err)
	}
	err := json.Unmarshal(raw, &tournament)
	return tournament, err
}

func (r *Repository) TournamentVisible(ctx context.Context, id uuid.UUID) (bool, error) {
	var visible bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tournaments WHERE id=$1 AND deleted_at IS NULL AND NOT hidden AND status<>'draft')`, id).Scan(&visible)
	return visible, dbError(err)
}

func (r *Repository) ListTournamentPage(ctx context.Context, p model.ListTournamentsParams, owner *uuid.UUID, private bool) (model.TournamentPage, error) {
	limit, offset := pagination.Values(p.Limit, p.Offset)
	where := ` FROM tournaments t JOIN users u ON u.id=t.organizer_id
 WHERE t.deleted_at IS NULL AND ($1::uuid IS NULL OR t.organizer_id=$1)
 AND ($2 OR (NOT t.hidden AND t.status<>'draft'))
 AND ($3::uuid IS NULL OR t.discipline_id=$3) AND ($4::text IS NULL OR t.status=$4)
 AND ($5::text IS NULL OR t.format=$5) AND ($6::timestamptz IS NULL OR t.starts_at >= $6)
 AND ($7::timestamptz IS NULL OR t.starts_at <= $7) AND ($8::text IS NULL OR t.name ILIKE '%'||$8||'%')`
	args := []any{owner, private, p.DisciplineId, p.Status, p.Format, p.StartsFrom, p.StartsTo, p.Search}
	page := model.TournamentPage{Items: []model.Tournament{}, Limit: limit, Offset: offset}
	if err := r.db.QueryRow(ctx, `SELECT count(*)`+where, args...).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+tournamentJSON+where+` ORDER BY t.created_at DESC,t.id ASC LIMIT $9 OFFSET $10`, append(args, limit, offset)...)
	if err != nil {
		return page, dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		var item model.Tournament
		if err := rows.Scan(&raw); err != nil {
			return page, dbError(err)
		}
		if err := json.Unmarshal(raw, &item); err != nil {
			return page, err
		}
		page.Items = append(page.Items, item)
	}
	return page, dbError(rows.Err())
}

func (r *Repository) ReadDisciplineRosterRules(ctx context.Context, id uuid.UUID) (repository.RosterRules, error) {
	var rules repository.RosterRules
	err := r.db.QueryRow(ctx, `SELECT min_roster_size,max_roster_size,allows_draw,score_unit FROM disciplines WHERE id=$1 AND active`, id).Scan(&rules.Min, &rules.Max, &rules.Draw, &rules.Unit)
	return rules, dbError(err)
}

func (r *Repository) ReadTournamentRosterRules(ctx context.Context, id uuid.UUID) (repository.RosterRules, error) {
	var rules repository.RosterRules
	err := r.db.QueryRow(ctx, `SELECT min_roster_size,max_roster_size,allows_draw,score_unit FROM tournaments WHERE id=$1`, id).Scan(&rules.Min, &rules.Max, &rules.Draw, &rules.Unit)
	return rules, dbError(err)
}

func tournamentRuleArgs(r model.Rules) []any {
	var count, qualifiers, wildcardPlace *int
	wildcard := 0
	if g := r.GroupSettings; g != nil {
		count, qualifiers, wildcardPlace, wildcard = &g.GroupCount, &g.QualifiersPerGroup, g.WildcardPlace, g.WildcardCount
	}
	ties := make([]string, len(r.TieBreakers))
	for i, tie := range r.TieBreakers {
		ties[i] = string(tie)
	}
	return []any{r.BestOf, r.RoundRobinLegs, r.WinPoints, r.DrawPoints, r.LossPoints, ties, r.ThirdPlaceMatch, r.ConfirmationTimeoutHours, count, qualifiers, wildcard, wildcardPlace}
}

func (r *Repository) CreateTournamentRecord(ctx context.Context, id, organizer uuid.UUID, request model.CreateTournamentRequest, roster repository.RosterRules) error {
	args := []any{id, organizer, request.DisciplineId, strings.TrimSpace(request.Name), request.Description, request.Format, request.MaxParticipants, request.RegistrationOpensAt, request.RegistrationClosesAt, request.StartsAt, request.EndsAt, request.Timezone}
	args = append(args, tournamentRuleArgs(request.Rules)...)
	args = append(args, roster.Min, roster.Max, roster.Draw, roster.Unit)
	_, err := r.db.Exec(ctx, `INSERT INTO tournaments(id,organizer_id,discipline_id,name,description,format,max_participants,registration_opens_at,registration_closes_at,starts_at,ends_at,timezone,best_of,round_robin_legs,win_points,draw_points,loss_points,tie_breakers,third_place_match,confirmation_timeout_hours,group_count,qualifiers_per_group,wildcard_count,wildcard_place,min_roster_size,max_roster_size,allows_draw,score_unit)
 VALUES($1,$2,$3,$4,COALESCE($5,''),$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28)`, args...)
	return dbError(err)
}

func (r *Repository) UpdateTournamentRecord(ctx context.Context, id uuid.UUID, request model.CreateTournamentRequest, roster repository.RosterRules) error {
	args := []any{id, request.DisciplineId, strings.TrimSpace(request.Name), request.Description, request.Format, request.MaxParticipants, request.RegistrationOpensAt, request.RegistrationClosesAt, request.StartsAt, request.EndsAt, request.Timezone}
	args = append(args, tournamentRuleArgs(request.Rules)...)
	args = append(args, roster.Min, roster.Max, roster.Draw, roster.Unit)
	_, err := r.db.Exec(ctx, `UPDATE tournaments SET discipline_id=$2,name=$3,description=COALESCE($4,''),format=$5,max_participants=$6,registration_opens_at=$7,registration_closes_at=$8,starts_at=$9,ends_at=$10,timezone=$11,best_of=$12,round_robin_legs=$13,win_points=$14,draw_points=$15,loss_points=$16,tie_breakers=$17,third_place_match=$18,confirmation_timeout_hours=$19,group_count=$20,qualifiers_per_group=$21,wildcard_count=$22,wildcard_place=$23,min_roster_size=$24,max_roster_size=$25,allows_draw=$26,score_unit=$27 WHERE id=$1`, args...)
	return dbError(err)
}

func (r *Repository) TournamentRosterCompatible(ctx context.Context, id uuid.UUID, minRoster, maxRoster int) (bool, error) {
	var incompatible bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM applications a WHERE a.tournament_id=$1 AND a.status IN ('pending','accepted') AND (SELECT count(*) FROM application_players p WHERE p.application_id=a.id) NOT BETWEEN $2 AND $3)`, id, minRoster, maxRoster).Scan(&incompatible)
	return !incompatible, dbError(err)
}

func (r *Repository) BumpTournamentVersion(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE tournaments SET version=version+1,updated_at=now() WHERE id=$1`, id)
	return dbError(err)
}

func (r *Repository) SetTournamentStatus(ctx context.Context, id uuid.UUID, status model.TournamentStatus) error {
	_, err := r.db.Exec(ctx, `UPDATE tournaments SET status=$2 WHERE id=$1`, id, status)
	return dbError(err)
}

func (r *Repository) SetTournamentReferee(ctx context.Context, id, referee uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE tournaments SET referee_id=$2 WHERE id=$1`, id, referee)
	return dbError(err)
}

func (r *Repository) TournamentPlayerIDs(ctx context.Context, id uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `SELECT DISTINCT ap.user_id FROM application_players ap JOIN applications a ON a.id=ap.application_id WHERE ap.tournament_id=$1 AND a.status IN ('pending','accepted') ORDER BY ap.user_id`, id)
	if err != nil {
		return nil, dbError(err)
	}
	ids, err := pgx.CollectRows(rows, pgx.RowTo[uuid.UUID])
	return ids, dbError(err)
}

func (r *Repository) MatchesUsingTournamentReferee(ctx context.Context, id uuid.UUID) ([]model.Match, error) {
	// Материализация остаётся внутри SQL-адаптера; сервис получает готовые модели.
	rows, err := r.db.Query(ctx, `SELECT to_jsonb(m) FROM matches m WHERE m.tournament_id=$1 AND m.referee_id IS NULL AND m.starts_at IS NOT NULL AND m.status NOT IN ('completed','bye') ORDER BY m.id`, id)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	matches := make([]model.Match, 0)
	for rows.Next() {
		var raw []byte
		var match model.Match
		if err := rows.Scan(&raw); err != nil {
			return nil, dbError(err)
		}
		if err := json.Unmarshal(raw, &match); err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}
	return matches, dbError(rows.Err())
}
