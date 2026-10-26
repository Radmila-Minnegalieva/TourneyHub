-- +goose Up
CREATE TABLE access_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    refresh_session_id uuid NOT NULL REFERENCES refresh_sessions(id),
    token_hash text NOT NULL UNIQUE,
    expires_at timestamptz NOT NULL,
    revoked_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX access_sessions_user_idx ON access_sessions(user_id);
CREATE INDEX access_sessions_refresh_idx ON access_sessions(refresh_session_id);
CREATE TABLE auth_rate_limits (
    ip text PRIMARY KEY,
    window_start timestamptz NOT NULL,
    attempts integer NOT NULL CHECK (attempts > 0)
);
-- Файлы до 5 МиБ хранятся вместе с метаданными и резервной копией БД.
ALTER TABLE files ADD COLUMN data bytea NOT NULL DEFAULT ''::bytea;
-- Сторона автора сохраняется даже после передачи капитанства или выхода из команды.
ALTER TABLE result_reports ADD COLUMN reporting_application_id uuid REFERENCES applications(id);
CREATE INDEX tournament_events_created_idx ON tournament_events(created_at);

-- +goose Down
DROP INDEX tournament_events_created_idx;
ALTER TABLE result_reports DROP COLUMN reporting_application_id;
ALTER TABLE files DROP COLUMN data;
DROP TABLE auth_rate_limits;
DROP TABLE access_sessions;
