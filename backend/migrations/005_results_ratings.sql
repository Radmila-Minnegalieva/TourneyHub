-- +goose Up
CREATE TABLE result_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id uuid NOT NULL REFERENCES matches(id),
    author_id uuid NOT NULL REFERENCES users(id),
    source text NOT NULL CHECK (source IN ('captain', 'referee', 'dispute')),
    home_score integer NOT NULL CHECK (home_score >= 0),
    away_score integer NOT NULL CHECK (away_score >= 0),
    status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'disputed', 'superseded')),
    confirmation_deadline timestamptz,
    confirmed_by uuid REFERENCES users(id),
    confirmed_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (id, match_id)
);
CREATE UNIQUE INDEX result_reports_pending_unique ON result_reports(match_id) WHERE status = 'pending';
CREATE INDEX result_reports_match_idx ON result_reports(match_id, created_at);
CREATE INDEX result_reports_deadline_idx ON result_reports(confirmation_deadline) WHERE status = 'pending';
CREATE TABLE report_games (
    report_id uuid NOT NULL REFERENCES result_reports(id),
    number integer NOT NULL CHECK (number BETWEEN 1 AND 5),
    home_score integer NOT NULL CHECK (home_score >= 0),
    away_score integer NOT NULL CHECK (away_score >= 0),
    PRIMARY KEY (report_id, number)
);
CREATE TABLE report_files (
    report_id uuid NOT NULL REFERENCES result_reports(id),
    file_id uuid NOT NULL REFERENCES files(id),
    PRIMARY KEY (report_id, file_id)
);
ALTER TABLE matches ADD COLUMN final_report_id uuid;
ALTER TABLE matches ADD CONSTRAINT matches_final_report_fk
    FOREIGN KEY (final_report_id, id) REFERENCES result_reports(id, match_id);
CREATE TABLE disputes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id uuid NOT NULL UNIQUE REFERENCES matches(id),
    report_id uuid NOT NULL,
    opened_by uuid NOT NULL REFERENCES users(id),
    reason text NOT NULL CHECK (length(reason) BETWEEN 1 AND 2000),
    status text NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'resolved')),
    resolved_by uuid REFERENCES users(id),
    resolution text,
    resolved_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    FOREIGN KEY (report_id, match_id) REFERENCES result_reports(id, match_id),
    CHECK (status <> 'resolved' OR
        (resolved_by IS NOT NULL AND resolved_at IS NOT NULL AND resolution IS NOT NULL AND length(resolution) > 0))
);
CREATE TABLE dispute_files (
    dispute_id uuid NOT NULL REFERENCES disputes(id),
    file_id uuid NOT NULL REFERENCES files(id),
    PRIMARY KEY (dispute_id, file_id)
);
CREATE TABLE ratings (
    team_id uuid NOT NULL REFERENCES teams(id),
    discipline_id uuid NOT NULL REFERENCES disciplines(id),
    value double precision NOT NULL DEFAULT 1500 CHECK (value NOT IN ('Infinity', '-Infinity', 'NaN')),
    played integer NOT NULL DEFAULT 0 CHECK (played >= 0),
    version bigint NOT NULL DEFAULT 1 CHECK (version > 0),
    PRIMARY KEY (team_id, discipline_id)
);
CREATE INDEX ratings_leaderboard_idx ON ratings(discipline_id, value DESC, team_id);
CREATE TABLE rating_changes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id uuid NOT NULL,
    tournament_id uuid NOT NULL,
    team_id uuid NOT NULL,
    discipline_id uuid NOT NULL,
    before double precision NOT NULL,
    delta double precision NOT NULL CHECK (delta NOT IN ('Infinity', '-Infinity', 'NaN')),
    after double precision NOT NULL,
    revision integer NOT NULL DEFAULT 1 CHECK (revision > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (match_id, team_id, revision),
    FOREIGN KEY (match_id, tournament_id) REFERENCES matches(id, tournament_id),
    FOREIGN KEY (tournament_id, discipline_id) REFERENCES tournaments(id, discipline_id),
    FOREIGN KEY (team_id, discipline_id) REFERENCES ratings(team_id, discipline_id),
    CHECK (abs(after - before - delta) < 0.000000001)
);
CREATE INDEX rating_changes_history_idx ON rating_changes(team_id, discipline_id, created_at DESC, id);

-- +goose Down
DROP TABLE rating_changes;
DROP TABLE ratings;
DROP TABLE dispute_files;
DROP TABLE disputes;
ALTER TABLE matches DROP CONSTRAINT matches_final_report_fk;
ALTER TABLE matches DROP COLUMN final_report_id;
DROP TABLE report_files;
DROP TABLE report_games;
DROP TABLE result_reports;
