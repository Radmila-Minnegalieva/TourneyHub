-- +goose Up
CREATE TABLE draw_previews (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id uuid NOT NULL REFERENCES tournaments(id),
    tournament_version bigint NOT NULL CHECK (tournament_version > 0),
    created_by uuid NOT NULL REFERENCES users(id),
    method text NOT NULL CHECK (method IN ('elo', 'random', 'manual')),
    payload jsonb NOT NULL CHECK (jsonb_typeof(payload) = 'object'),
    expires_at timestamptz NOT NULL,
    confirmed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (expires_at > created_at)
);
CREATE INDEX draw_previews_tournament_idx ON draw_previews(tournament_id);
CREATE TABLE stages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id uuid NOT NULL REFERENCES tournaments(id),
    kind text NOT NULL CHECK (kind IN ('round_robin', 'groups', 'playoffs')),
    position integer NOT NULL CHECK (position > 0),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed')),
    UNIQUE (id, tournament_id),
    UNIQUE (tournament_id, position)
);
CREATE TABLE tournament_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    stage_id uuid NOT NULL,
    tournament_id uuid NOT NULL,
    name text NOT NULL,
    UNIQUE (id, stage_id, tournament_id),
    UNIQUE (stage_id, name),
    FOREIGN KEY (stage_id, tournament_id) REFERENCES stages(id, tournament_id)
);
CREATE TABLE group_participants (
    group_id uuid NOT NULL,
    stage_id uuid NOT NULL,
    tournament_id uuid NOT NULL,
    application_id uuid NOT NULL,
    PRIMARY KEY (group_id, application_id),
    UNIQUE (stage_id, application_id),
    FOREIGN KEY (group_id, stage_id, tournament_id) REFERENCES tournament_groups(id, stage_id, tournament_id),
    FOREIGN KEY (application_id, tournament_id) REFERENCES applications(id, tournament_id)
);
CREATE TABLE matches (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tournament_id uuid NOT NULL,
    stage_id uuid NOT NULL,
    group_id uuid,
    round integer NOT NULL CHECK (round > 0),
    position integer NOT NULL CHECK (position > 0),
    home_application_id uuid,
    away_application_id uuid,
    winner_application_id uuid,
    next_match_id uuid,
    next_match_slot text CHECK (next_match_slot IN ('home', 'away')),
    loser_next_match_id uuid,
    loser_next_match_slot text CHECK (loser_next_match_slot IN ('home', 'away')),
    starts_at timestamptz,
    venue text CHECK (length(venue) <= 200),
    online_url text,
    referee_id uuid REFERENCES users(id),
    status text NOT NULL DEFAULT 'scheduled'
        CHECK (status IN ('scheduled', 'awaiting_confirmation', 'disputed', 'completed', 'bye')),
    home_score integer CHECK (home_score >= 0),
    away_score integer CHECK (away_score >= 0),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (id, tournament_id),
    UNIQUE (stage_id, round, position),
    FOREIGN KEY (stage_id, tournament_id) REFERENCES stages(id, tournament_id),
    FOREIGN KEY (group_id, stage_id, tournament_id) REFERENCES tournament_groups(id, stage_id, tournament_id),
    FOREIGN KEY (home_application_id, tournament_id) REFERENCES applications(id, tournament_id),
    FOREIGN KEY (away_application_id, tournament_id) REFERENCES applications(id, tournament_id),
    FOREIGN KEY (winner_application_id, tournament_id) REFERENCES applications(id, tournament_id),
    FOREIGN KEY (next_match_id, tournament_id) REFERENCES matches(id, tournament_id) DEFERRABLE INITIALLY DEFERRED,
    FOREIGN KEY (loser_next_match_id, tournament_id) REFERENCES matches(id, tournament_id) DEFERRABLE INITIALLY DEFERRED,
    CHECK (home_application_id IS NULL OR away_application_id IS NULL OR home_application_id <> away_application_id),
    CHECK (winner_application_id IS NULL OR
        (home_application_id IS NOT NULL AND winner_application_id = home_application_id) OR
        (away_application_id IS NOT NULL AND winner_application_id = away_application_id)),
    CHECK ((next_match_id IS NULL) = (next_match_slot IS NULL)),
    CHECK ((loser_next_match_id IS NULL) = (loser_next_match_slot IS NULL)),
    CHECK (next_match_id IS NULL OR next_match_id <> id),
    CHECK (loser_next_match_id IS NULL OR loser_next_match_id <> id),
    CHECK (status <> 'completed' OR
        (home_application_id IS NOT NULL AND away_application_id IS NOT NULL
         AND home_score IS NOT NULL AND away_score IS NOT NULL))
);
CREATE INDEX matches_schedule_idx ON matches(tournament_id, starts_at, id);
CREATE INDEX matches_referee_idx ON matches(referee_id, status);
CREATE TABLE standings (
    stage_id uuid NOT NULL,
    tournament_id uuid NOT NULL,
    group_id uuid,
    application_id uuid NOT NULL,
    position integer NOT NULL CHECK (position > 0),
    played integer NOT NULL DEFAULT 0 CHECK (played >= 0),
    wins integer NOT NULL DEFAULT 0 CHECK (wins >= 0),
    draws integer NOT NULL DEFAULT 0 CHECK (draws >= 0),
    losses integer NOT NULL DEFAULT 0 CHECK (losses >= 0),
    score_for integer NOT NULL DEFAULT 0 CHECK (score_for >= 0),
    score_against integer NOT NULL DEFAULT 0 CHECK (score_against >= 0),
    points integer NOT NULL DEFAULT 0,
    PRIMARY KEY (stage_id, application_id),
    FOREIGN KEY (stage_id, tournament_id) REFERENCES stages(id, tournament_id),
    FOREIGN KEY (group_id, stage_id, tournament_id) REFERENCES tournament_groups(id, stage_id, tournament_id),
    FOREIGN KEY (application_id, tournament_id) REFERENCES applications(id, tournament_id),
    CHECK (played = wins + draws + losses)
);

-- +goose Down
DROP TABLE standings;
DROP TABLE matches;
DROP TABLE group_participants;
DROP TABLE tournament_groups;
DROP TABLE stages;
DROP TABLE draw_previews;
