-- +goose Up
CREATE TABLE teams (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 120),
    description text NOT NULL DEFAULT '' CHECK (length(description) <= 2000),
    logo_file_id uuid REFERENCES files(id),
    captain_id uuid NOT NULL REFERENCES users(id),
    personal boolean NOT NULL DEFAULT false,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    deleted_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX teams_personal_owner_unique ON teams(captain_id) WHERE personal AND deleted_at IS NULL;
CREATE TABLE team_members (
    team_id uuid NOT NULL REFERENCES teams(id),
    user_id uuid NOT NULL REFERENCES users(id),
    joined_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (team_id, user_id)
);
-- Команда и членство капитана создаются или изменяются в одной транзакции.
ALTER TABLE teams ADD CONSTRAINT teams_captain_membership_fk
    FOREIGN KEY (id, captain_id) REFERENCES team_members(team_id, user_id)
    DEFERRABLE INITIALLY DEFERRED;
CREATE INDEX team_members_user_idx ON team_members(user_id);

CREATE TABLE invitations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id uuid NOT NULL REFERENCES teams(id),
    invitee_id uuid REFERENCES users(id),
    created_by uuid NOT NULL REFERENCES users(id),
    token_hash text NOT NULL UNIQUE,
    status text NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending', 'accepted', 'declined', 'revoked', 'expired')),
    consumed_by uuid REFERENCES users(id),
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at)
);
CREATE INDEX invitations_invitee_idx ON invitations(invitee_id, status);
CREATE INDEX invitations_team_idx ON invitations(team_id);

-- +goose Down
DROP TABLE invitations;
ALTER TABLE teams DROP CONSTRAINT teams_captain_membership_fk;
DROP TABLE team_members;
DROP TABLE teams;
