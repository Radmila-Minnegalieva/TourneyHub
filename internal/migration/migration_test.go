//go:build integration

package migration_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Radmila-Minnegalieva/TourneyHub/internal/migration"
	"github.com/Radmila-Minnegalieva/TourneyHub/migrations"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestMigrationLifecycleAndConstraints(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Fatal("TEST_DATABASE_URL is required; use a separate empty test database")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	provider, db, err := migration.Open(ctx, dsn, migrations.Files)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	// Refuse a database containing application tables, even if someone supplied the wrong DSN.
	var count int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema='public' AND table_name <> 'goose_db_version'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("test requires an empty database; refusing to roll back existing tables")
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	if results, err := provider.Up(ctx); err != nil || len(results) != 0 {
		t.Fatalf("second up: %v %v", results, err)
	}
	var tableCount int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM information_schema.tables WHERE table_schema='public'`).Scan(&tableCount); err != nil {
		t.Fatal(err)
	}
	if tableCount < 25 {
		t.Fatalf("missing domain tables: %d", tableCount)
	}
	// Verify important DB constraints with real SQL; every probe is rolled back.
	for _, tc := range []struct{ name, sql, code string }{
		{"file size", `INSERT INTO files(owner_id,purpose,storage_key,filename,content_type,size_bytes) VALUES ('00000000-0000-0000-0000-000000000001','match_protocol','oversize','test.pdf','application/pdf',5242881)`, "23514"},
		{"duplicate nickname", `INSERT INTO users(id,email,password_hash,nickname,personal_data_consent_at) VALUES ('00000000-0000-0000-0000-000000000002','other@example.com','hash','PLAYER',now())`, "23505"},
		{"captain outside team", `INSERT INTO teams(name,captain_id) VALUES ('orphan','00000000-0000-0000-0000-000000000001')`, "23503"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err = tx.ExecContext(ctx, `INSERT INTO users(id,email,password_hash,nickname,personal_data_consent_at) VALUES ('00000000-0000-0000-0000-000000000001','player@example.com','hash','player',now())`); err != nil {
				t.Fatal(err)
			}
			_, err = tx.ExecContext(ctx, tc.sql)
			if err == nil {
				_, err = tx.ExecContext(ctx, "SET CONSTRAINTS ALL IMMEDIATE")
			}
			if err == nil {
				t.Fatal("constraint accepted invalid data")
			}
			var pgError *pgconn.PgError
			if !errors.As(err, &pgError) || pgError.Code != tc.code {
				t.Fatalf("unexpected constraint error: %v", err)
			}
		})
	}
	t.Run("player cannot represent two teams in one tournament", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback()
		const fixture = `
INSERT INTO users(id,email,password_hash,nickname,personal_data_consent_at) VALUES
('10000000-0000-0000-0000-000000000001','a@example.com','hash','player_a',now()),
('10000000-0000-0000-0000-000000000002','b@example.com','hash','player_b',now());
INSERT INTO teams(id,name,captain_id) VALUES
('20000000-0000-0000-0000-000000000001','A','10000000-0000-0000-0000-000000000001'),
('20000000-0000-0000-0000-000000000002','B','10000000-0000-0000-0000-000000000002');
INSERT INTO team_members(team_id,user_id) VALUES
('20000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001'),
('20000000-0000-0000-0000-000000000002','10000000-0000-0000-0000-000000000001'),
('20000000-0000-0000-0000-000000000002','10000000-0000-0000-0000-000000000002');
INSERT INTO disciplines(id,name,min_roster_size,max_roster_size,allows_draw,score_unit)
VALUES ('30000000-0000-0000-0000-000000000001','chess',1,1,false,'points');
INSERT INTO tournaments(id,organizer_id,discipline_id,name,format,max_participants,best_of,win_points,draw_points,loss_points,tie_breakers,min_roster_size,max_roster_size,allows_draw,score_unit)
VALUES ('40000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001','30000000-0000-0000-0000-000000000001','Cup','single_elimination',2,1,3,1,0,ARRAY['head_to_head'],1,1,false,'points');
INSERT INTO applications(id,tournament_id,team_id,submitted_by) VALUES
('50000000-0000-0000-0000-000000000001','40000000-0000-0000-0000-000000000001','20000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001'),
('50000000-0000-0000-0000-000000000002','40000000-0000-0000-0000-000000000001','20000000-0000-0000-0000-000000000002','10000000-0000-0000-0000-000000000002');
INSERT INTO application_players(application_id,tournament_id,user_id) VALUES
('50000000-0000-0000-0000-000000000001','40000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001'),
('50000000-0000-0000-0000-000000000002','40000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001');
INSERT INTO tournament_player_claims(tournament_id,user_id,application_id)
VALUES ('40000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001','50000000-0000-0000-0000-000000000001');
SET CONSTRAINTS ALL IMMEDIATE;`
		if _, err := tx.ExecContext(ctx, fixture); err != nil {
			t.Fatal(err)
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO tournament_player_claims(tournament_id,user_id,application_id) VALUES ('40000000-0000-0000-0000-000000000001','10000000-0000-0000-0000-000000000001','50000000-0000-0000-0000-000000000002')`)
		var pgError *pgconn.PgError
		if !errors.As(err, &pgError) || pgError.Code != "23505" {
			t.Fatalf("expected duplicate player rejection, got %v", err)
		}
	})
	if t.Failed() {
		return
	}
	if _, err := provider.DownTo(ctx, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	current, target, err := provider.GetVersions(ctx)
	if err != nil || current != 6 || target != 6 {
		t.Fatal(fmt.Sprintf("versions %d/%d: %v", current, target, err))
	}
	if _, err := provider.DownTo(ctx, 0); err != nil {
		t.Fatal(err)
	}
}
