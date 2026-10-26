package postgres

import (
	"context"

	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/model"
	"github.com/Radmila-Minnegalieva/TourneyHub/backend/internal/repository"
	"github.com/google/uuid"
)

const teamColumns = `t.id,t.name,t.description,t.captain_id,t.personal,t.version,
 CASE WHEN t.logo_file_id IS NULL THEN NULL ELSE '/api/v1/files/'||t.logo_file_id END`

func scanTeam(row identityScanner) (model.Team, error) {
	var team model.Team
	err := row.Scan(&team.Id, &team.Name, &team.Description, &team.CaptainId, &team.Personal, &team.Version, &team.LogoUrl)
	return team, dbError(err)
}

func (r *Repository) Team(ctx context.Context, id uuid.UUID, lock bool) (model.Team, error) {
	query := `SELECT ` + teamColumns + ` FROM teams t WHERE t.id=$1 AND t.deleted_at IS NULL`
	if lock {
		query += ` FOR UPDATE`
	}
	return scanTeam(r.db.QueryRow(ctx, query, id))
}

func (r *Repository) LockUserNickname(ctx context.Context, id uuid.UUID) (string, error) {
	var nickname string
	err := r.db.QueryRow(ctx, `SELECT nickname FROM users WHERE id=$1 FOR UPDATE`, id).Scan(&nickname)
	return nickname, dbError(err)
}

func (r *Repository) PersonalTeam(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM teams WHERE captain_id=$1 AND personal AND deleted_at IS NULL`, userID).Scan(&id)
	return id, dbError(err)
}

func (r *Repository) CreateTeamRecord(ctx context.Context, name string, description *string, logoID *uuid.UUID, captainID uuid.UUID, personal bool) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `INSERT INTO teams(name,description,logo_file_id,captain_id,personal) VALUES($1,COALESCE($2,''),$3,$4,$5) RETURNING id`, name, description, logoID, captainID, personal).Scan(&id)
	return id, dbError(err)
}

func (r *Repository) AddTeamMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	tag, err := r.db.Exec(ctx, `INSERT INTO team_members(team_id,user_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, teamID, userID)
	return tag.RowsAffected() > 0, dbError(err)
}

func (r *Repository) TeamsPage(ctx context.Context, search string, owner *uuid.UUID, limit, offset int) (model.TeamPage, error) {
	page := model.TeamPage{Items: []model.Team{}, Limit: limit, Offset: offset}
	where := `t.deleted_at IS NULL AND ($1='' OR strpos(lower(t.name),lower($1))>0) AND ($2::uuid IS NULL OR EXISTS(SELECT 1 FROM team_members tm WHERE tm.team_id=t.id AND tm.user_id=$2))`
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM teams t WHERE `+where, search, owner).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	rows, err := r.db.Query(ctx, `SELECT `+teamColumns+` FROM teams t WHERE `+where+` ORDER BY t.created_at DESC,t.id LIMIT $3 OFFSET $4`, search, owner, limit, offset)
	if err != nil {
		return page, dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		team, err := scanTeam(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, team)
	}
	return page, dbError(rows.Err())
}

func (r *Repository) UpdateTeamRecord(ctx context.Context, id uuid.UUID, req model.UpdateTeamRequest) error {
	_, err := r.db.Exec(ctx, `UPDATE teams SET name=COALESCE($2,name),description=COALESCE($3,description),logo_file_id=COALESCE($4,logo_file_id),version=version+1,updated_at=now() WHERE id=$1`, id, req.Name, req.Description, req.LogoFileId)
	return dbError(err)
}

func (r *Repository) TeamHasActiveApplications(ctx context.Context, id uuid.UUID) (bool, error) {
	var active bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM applications a JOIN tournaments t ON t.id=a.tournament_id WHERE a.team_id=$1 AND a.status='accepted' AND t.status IN ('registration','in_progress') AND t.deleted_at IS NULL)`, id).Scan(&active)
	return active, dbError(err)
}

func (r *Repository) WithdrawPendingTeamApplications(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM tournament_player_claims WHERE application_id IN (SELECT id FROM applications WHERE team_id=$1 AND status='pending')`, id); err != nil {
		return dbError(err)
	}
	_, err := r.db.Exec(ctx, `UPDATE applications SET status='withdrawn',version=version+1,updated_at=now() WHERE team_id=$1 AND status='pending'`, id)
	return dbError(err)
}

func (r *Repository) RevokeTeamInvitations(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE invitations SET status='revoked' WHERE team_id=$1 AND status='pending'`, id)
	return dbError(err)
}

func (r *Repository) SoftDeleteTeam(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE teams SET deleted_at=now(),version=version+1,updated_at=now() WHERE id=$1`, id)
	return dbError(err)
}

func (r *Repository) TeamMembers(ctx context.Context, id uuid.UUID) ([]model.TeamMember, error) {
	rows, err := r.db.Query(ctx, `SELECT u.id,u.nickname,u.name,CASE WHEN u.avatar_file_id IS NULL THEN NULL ELSE '/api/v1/files/'||u.avatar_file_id END,tm.joined_at,u.id=t.captain_id FROM team_members tm JOIN users u ON u.id=tm.user_id JOIN teams t ON t.id=tm.team_id WHERE tm.team_id=$1 ORDER BY u.id=t.captain_id DESC,tm.joined_at,u.id`, id)
	if err != nil {
		return nil, dbError(err)
	}
	defer rows.Close()
	items := make([]model.TeamMember, 0)
	for rows.Next() {
		var member model.TeamMember
		if err := rows.Scan(&member.User.Id, &member.User.Nickname, &member.User.Name, &member.User.AvatarUrl, &member.JoinedAt, &member.Captain); err != nil {
			return nil, dbError(err)
		}
		items = append(items, member)
	}
	return items, dbError(rows.Err())
}

func (r *Repository) DeleteTeamMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM team_members WHERE team_id=$1 AND user_id=$2`, teamID, userID)
	return tag.RowsAffected() > 0, dbError(err)
}

func (r *Repository) BumpTeamVersion(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE teams SET version=version+1,updated_at=now() WHERE id=$1`, id)
	return dbError(err)
}

func (r *Repository) ActiveTeamMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	var active bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM team_members tm JOIN users u ON u.id=tm.user_id WHERE tm.team_id=$1 AND tm.user_id=$2 AND u.blocked_at IS NULL)`, teamID, userID).Scan(&active)
	return active, dbError(err)
}

func (r *Repository) TransferTeamCaptain(ctx context.Context, teamID, userID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE teams SET captain_id=$2,version=version+1,updated_at=now() WHERE id=$1`, teamID, userID)
	return dbError(err)
}

func (r *Repository) InviteeByNickname(ctx context.Context, nickname string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM users WHERE lower(nickname)=lower($1) AND blocked_at IS NULL AND email_verified_at IS NOT NULL`, nickname).Scan(&id)
	return id, dbError(err)
}

func (r *Repository) IsTeamMember(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	var member bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM team_members WHERE team_id=$1 AND user_id=$2)`, teamID, userID).Scan(&member)
	return member, dbError(err)
}

func (r *Repository) CreateInvitationRecord(ctx context.Context, teamID uuid.UUID, inviteeID *uuid.UUID, actor uuid.UUID, hash string) (model.Invitation, error) {
	invitation := model.Invitation{TeamId: teamID, InviteeId: inviteeID, Status: model.InvitationStatusPending}
	err := r.db.QueryRow(ctx, `INSERT INTO invitations(team_id,invitee_id,created_by,token_hash,expires_at) VALUES($1,$2,$3,$4,now()+interval '7 days') RETURNING id,expires_at`, teamID, inviteeID, actor, hash).Scan(&invitation.Id, &invitation.ExpiresAt)
	return invitation, dbError(err)
}

func scanInvitation(row identityScanner) (model.Invitation, error) {
	var invitation model.Invitation
	err := row.Scan(&invitation.Id, &invitation.TeamId, &invitation.InviteeId, &invitation.Status, &invitation.ExpiresAt)
	return invitation, dbError(err)
}

func (r *Repository) InvitationsPage(ctx context.Context, actor uuid.UUID, limit, offset int) (model.InvitationPage, error) {
	page := model.InvitationPage{Items: []model.Invitation{}, Limit: limit, Offset: offset}
	where := `(invitee_id=$1 OR consumed_by=$1)`
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM invitations WHERE `+where, actor).Scan(&page.Total); err != nil {
		return page, dbError(err)
	}
	rows, err := r.db.Query(ctx, `SELECT id,team_id,invitee_id,CASE WHEN status='pending' AND expires_at<=now() THEN 'expired' ELSE status END,expires_at FROM invitations WHERE `+where+` ORDER BY created_at DESC,id LIMIT $2 OFFSET $3`, actor, limit, offset)
	if err != nil {
		return page, dbError(err)
	}
	defer rows.Close()
	for rows.Next() {
		invitation, err := scanInvitation(rows)
		if err != nil {
			return page, err
		}
		page.Items = append(page.Items, invitation)
	}
	return page, dbError(rows.Err())
}

func (r *Repository) InvitationTeam(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	var teamID uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT team_id FROM invitations WHERE id=$1`, id).Scan(&teamID)
	return teamID, dbError(err)
}

func (r *Repository) LockInvitation(ctx context.Context, id uuid.UUID) (repository.InvitationSecret, error) {
	var record repository.InvitationSecret
	i := &record.Invitation
	err := r.db.QueryRow(ctx, `SELECT id,team_id,invitee_id,status,expires_at,token_hash FROM invitations WHERE id=$1 FOR UPDATE`, id).Scan(&i.Id, &i.TeamId, &i.InviteeId, &i.Status, &i.ExpiresAt, &record.TokenHash)
	return record, dbError(err)
}

func (r *Repository) SetInvitationDecision(ctx context.Context, id uuid.UUID, status model.InvitationStatus, actor uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE invitations SET status=$2,consumed_by=$3 WHERE id=$1`, id, string(status), actor)
	return dbError(err)
}

func (r *Repository) LockTeamInvitationStatus(ctx context.Context, id, teamID uuid.UUID) (string, error) {
	var status string
	err := r.db.QueryRow(ctx, `SELECT status FROM invitations WHERE id=$1 AND team_id=$2 FOR UPDATE`, id, teamID).Scan(&status)
	return status, dbError(err)
}

func (r *Repository) RevokeInvitationRecord(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE invitations SET status='revoked' WHERE id=$1`, id)
	return dbError(err)
}
