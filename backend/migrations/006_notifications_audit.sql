-- +goose Up
CREATE TABLE notifications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id),
    kind text NOT NULL,
    message text NOT NULL,
    resource_id uuid,
    read_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notifications_user_idx ON notifications(user_id, created_at DESC, id);
CREATE TABLE audit_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_id uuid REFERENCES users(id),
    action text NOT NULL,
    entity_type text NOT NULL,
    entity_id uuid NOT NULL,
    before jsonb,
    after jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_entries_entity_idx ON audit_entries(entity_id, created_at DESC, id);
CREATE INDEX audit_entries_actor_idx ON audit_entries(actor_id, created_at DESC, id);
-- Written with the domain transaction; future workers deliver SMTP/SSE after commit.
CREATE TABLE outbox_events (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    topic text NOT NULL,
    aggregate_id uuid NOT NULL,
    payload jsonb NOT NULL,
    attempts integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at timestamptz NOT NULL DEFAULT now(),
    delivered_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX outbox_events_pending_idx ON outbox_events(available_at, id) WHERE delivered_at IS NULL;
CREATE TABLE tournament_events (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tournament_id uuid NOT NULL REFERENCES tournaments(id),
    kind text NOT NULL CHECK (kind IN ('match_updated', 'standings_updated', 'bracket_updated', 'tournament_updated')),
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tournament_events_replay_idx ON tournament_events(tournament_id, id);

-- +goose Down
DROP TABLE tournament_events;
DROP TABLE outbox_events;
DROP TABLE audit_entries;
DROP TABLE notifications;
