-- +goose Up
CREATE TABLE disciplines (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 120),
    min_roster_size integer NOT NULL CHECK (min_roster_size BETWEEN 1 AND 100),
    max_roster_size integer NOT NULL CHECK (max_roster_size BETWEEN 1 AND 100),
    allows_draw boolean NOT NULL,
    score_unit text NOT NULL CHECK (length(score_unit) BETWEEN 1 AND 40),
    active boolean NOT NULL DEFAULT true,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    CHECK (min_roster_size <= max_roster_size)
);
CREATE UNIQUE INDEX disciplines_name_unique ON disciplines(lower(name));
CREATE TABLE tournaments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organizer_id uuid NOT NULL REFERENCES users(id),
    discipline_id uuid NOT NULL REFERENCES disciplines(id),
    referee_id uuid REFERENCES users(id),
    name text NOT NULL CHECK (length(name) BETWEEN 1 AND 120),
    description text NOT NULL DEFAULT '' CHECK (length(description) <= 10000),
    format text NOT NULL CHECK (format IN ('single_elimination', 'round_robin', 'groups_playoffs')),
    status text NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'registration', 'in_progress', 'completed', 'cancelled')),
    max_participants integer NOT NULL CHECK (max_participants BETWEEN 2 AND 256),
    registration_opens_at timestamptz,
    registration_closes_at timestamptz,
    starts_at timestamptz,
    ends_at timestamptz,
    timezone text NOT NULL DEFAULT 'UTC',
    best_of integer NOT NULL CHECK (best_of IN (1, 3, 5)),
    round_robin_legs integer NOT NULL DEFAULT 1 CHECK (round_robin_legs IN (1, 2)),
    win_points integer NOT NULL CHECK (win_points BETWEEN 0 AND 100),
    draw_points integer NOT NULL CHECK (draw_points BETWEEN 0 AND 100),
    loss_points integer NOT NULL CHECK (loss_points BETWEEN 0 AND 100),
    tie_breakers text[] NOT NULL CHECK (cardinality(tie_breakers) > 0
        AND tie_breakers <@ ARRAY['head_to_head', 'score_difference', 'score_for', 'elo', 'draw']),
    third_place_match boolean NOT NULL DEFAULT false,
    confirmation_timeout_hours integer NOT NULL DEFAULT 24 CHECK (confirmation_timeout_hours BETWEEN 1 AND 168),
    group_count integer CHECK (group_count BETWEEN 2 AND 64),
    qualifiers_per_group integer CHECK (qualifiers_per_group BETWEEN 1 AND 128),
    wildcard_count integer NOT NULL DEFAULT 0 CHECK (wildcard_count BETWEEN 0 AND 128),
    wildcard_place integer CHECK (wildcard_place BETWEEN 2 AND 128),
    -- Снимок регламента защищает активные турниры от изменения дисциплины.
    min_roster_size integer NOT NULL CHECK (min_roster_size BETWEEN 1 AND 100),
    max_roster_size integer NOT NULL CHECK (max_roster_size BETWEEN 1 AND 100),
    allows_draw boolean NOT NULL,
    score_unit text NOT NULL,
    hidden boolean NOT NULL DEFAULT false,
    deleted_at timestamptz,
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (id, discipline_id),
    CHECK (min_roster_size <= max_roster_size),
    CHECK (registration_opens_at < registration_closes_at),
    CHECK (registration_closes_at <= starts_at),
    CHECK (starts_at < ends_at),
    CHECK (status IN ('draft', 'cancelled') OR
        (registration_opens_at IS NOT NULL AND registration_closes_at IS NOT NULL
         AND starts_at IS NOT NULL AND ends_at IS NOT NULL)),
    CHECK ((format = 'groups_playoffs' AND group_count IS NOT NULL AND qualifiers_per_group IS NOT NULL)
        OR (format <> 'groups_playoffs' AND group_count IS NULL AND qualifiers_per_group IS NULL AND wildcard_count = 0)),
    CHECK (wildcard_count = 0 OR wildcard_place IS NOT NULL)
);
CREATE INDEX tournaments_catalog_idx ON tournaments(discipline_id, status, starts_at, id)
    WHERE NOT hidden AND deleted_at IS NULL AND status <> 'draft';
CREATE INDEX tournaments_organizer_idx ON tournaments(organizer_id, created_at DESC, id);
CREATE TABLE applications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id uuid NOT NULL REFERENCES tournaments(id),
    team_id uuid NOT NULL REFERENCES teams(id),
    submitted_by uuid NOT NULL REFERENCES users(id),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'withdrawn')),
    rejection_reason text,
    seed integer CHECK (seed BETWEEN 1 AND 256),
    final_place integer CHECK (final_place BETWEEN 1 AND 256),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (id, tournament_id),
    UNIQUE (id, tournament_id, team_id),
    CHECK (status <> 'rejected' OR (rejection_reason IS NOT NULL AND length(rejection_reason) > 0))
);
CREATE UNIQUE INDEX applications_active_team_unique ON applications(tournament_id, team_id)
    WHERE status IN ('pending', 'accepted');
CREATE UNIQUE INDEX applications_seed_unique ON applications(tournament_id, seed) WHERE status = 'accepted';
CREATE INDEX applications_tournament_status_idx ON applications(tournament_id, status);
CREATE INDEX applications_team_idx ON applications(team_id);
CREATE TABLE application_players (
    application_id uuid NOT NULL,
    tournament_id uuid NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id),
    PRIMARY KEY (application_id, user_id),
    UNIQUE (application_id, tournament_id, user_id),
    FOREIGN KEY (application_id, tournament_id) REFERENCES applications(id, tournament_id)
);
-- Резерв создаётся для pending/accepted и удаляется при rejected/withdrawn в той же транзакции.
-- Исторический снимок состава хранится отдельно и не резервирует игрока.
CREATE TABLE tournament_player_claims (
    tournament_id uuid NOT NULL,
    user_id uuid NOT NULL,
    application_id uuid NOT NULL,
    PRIMARY KEY (tournament_id, user_id),
    FOREIGN KEY (application_id, tournament_id, user_id)
        REFERENCES application_players(application_id, tournament_id, user_id)
);

-- +goose Down
DROP TABLE tournament_player_claims;
DROP TABLE application_players;
DROP TABLE applications;
DROP TABLE tournaments;
DROP TABLE disciplines;
