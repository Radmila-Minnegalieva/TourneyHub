-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text NOT NULL CHECK (length(email) <= 254),
    password_hash text NOT NULL,
    nickname text NOT NULL CHECK (nickname ~ '^[A-Za-z0-9_]{3,32}$'),
    name text CHECK (length(name) <= 120),
    timezone text NOT NULL DEFAULT 'UTC',
    email_notifications boolean NOT NULL DEFAULT true,
    email_verified_at timestamptz,
    blocked_at timestamptz,
    block_reason text,
    personal_data_consent_at timestamptz NOT NULL,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_unique ON users (lower(email));
CREATE UNIQUE INDEX users_nickname_unique ON users (lower(nickname));

CREATE TABLE user_roles (
    user_id uuid NOT NULL REFERENCES users(id),
    role text NOT NULL CHECK (role IN ('participant', 'organizer', 'referee', 'admin')),
    PRIMARY KEY (user_id, role)
);
CREATE TABLE refresh_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at)
);
CREATE INDEX refresh_sessions_user_idx ON refresh_sessions(user_id);
CREATE TABLE user_tokens (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    purpose text NOT NULL CHECK (purpose IN ('verify_email', 'reset_password')),
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    consumed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at)
);
CREATE INDEX user_tokens_user_idx ON user_tokens(user_id);

CREATE TABLE files (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id uuid NOT NULL REFERENCES users(id),
    purpose text NOT NULL CHECK (purpose IN ('avatar', 'team_logo', 'match_protocol')),
    storage_key text NOT NULL UNIQUE,
    filename text NOT NULL CHECK (length(filename) BETWEEN 1 AND 255),
    content_type text NOT NULL CHECK (content_type IN ('application/pdf', 'image/png', 'image/jpeg')),
    size_bytes integer NOT NULL CHECK (size_bytes BETWEEN 1 AND 5242880),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (purpose = 'match_protocol' OR content_type <> 'application/pdf')
);
CREATE INDEX files_owner_idx ON files(owner_id);
ALTER TABLE users ADD COLUMN avatar_file_id uuid REFERENCES files(id);

-- +goose Down
ALTER TABLE users DROP COLUMN avatar_file_id;
DROP TABLE files;
DROP TABLE user_tokens;
DROP TABLE refresh_sessions;
DROP TABLE user_roles;
DROP TABLE users;
