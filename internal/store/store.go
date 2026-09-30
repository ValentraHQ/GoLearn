// Package store persists learner state in PostgreSQL or SQLite.
package store

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/valentrahq/golearn/migrations"
)

type Store struct {
	db       *sql.DB
	postgres bool
}

// Open connects using a URL: "postgres://..." or "sqlite:path" (":memory:" allowed).
func Open(ctx context.Context, url string) (*Store, error) {
	var (
		db  *sql.DB
		err error
		pg  bool
	)
	switch {
	case strings.HasPrefix(url, "postgres://"), strings.HasPrefix(url, "postgresql://"):
		pg = true
		db, err = sql.Open("pgx", url)
	case strings.HasPrefix(url, "sqlite:"):
		path := strings.TrimPrefix(url, "sqlite:")
		dsn := "file:" + path + "?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
		if path == ":memory:" {
			dsn = "file::memory:?cache=shared&_pragma=foreign_keys(1)"
		}
		db, err = sql.Open("sqlite", dsn)
		if err == nil {
			db.SetMaxOpenConns(1) // serialise writers; avoids SQLITE_BUSY
		}
	default:
		return nil, fmt.Errorf("unsupported DATABASE_URL %q (use postgres:// or sqlite:)", url)
	}
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	s := &Store{db: db, postgres: pg}
	if err := s.migrate(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }

// q rewrites "?" placeholders to "$n" for PostgreSQL.
func (s *Store) q(query string) string {
	if !s.postgres {
		return query
	}
	var b strings.Builder
	n := 0
	for _, r := range query {
		if r == '?' {
			n++
			b.WriteString("$" + strconv.Itoa(n))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY, applied_at BIGINT NOT NULL)`); err != nil {
		return err
	}
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return err
	}
	names := []string{}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		var n int
		if err := s.db.QueryRowContext(ctx, s.q(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`), name).Scan(&n); err != nil {
			return err
		}
		if n > 0 {
			continue
		}
		body, err := migrations.FS.ReadFile(name)
		if err != nil {
			return err
		}
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("migration %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, s.q(`INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`), name, now()); err != nil {
			_ = tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}
