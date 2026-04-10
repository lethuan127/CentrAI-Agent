// Package sqlite implements [session.Store] with a single-file SQLite database (pure Go driver).
package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	_ "modernc.org/sqlite" // SQLite driver

	"github.com/lethuan127/centrai-agent/internal/session"
)

const schemaVersion = 1

// Store persists sessions as JSON blobs keyed by session id.
type Store struct {
	db *sql.DB
}

// Open opens (or creates) a SQLite file and applies migrations.
// The DSN enables a bounded busy wait so concurrent writes do not fail immediately.
func Open(path string) (*Store, error) {
	dsn := sqliteDSN(path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func sqliteDSN(path string) string {
	switch {
	case path == "", path == ":memory:":
		return path
	case strings.Contains(path, "?"):
		return path
	default:
		return path + "?_pragma=busy_timeout(5000)"
	}
}

func (s *Store) migrate() error {
	if _, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS centrai_schema_version (version INTEGER NOT NULL)`); err != nil {
		return fmt.Errorf("sqlite: schema table: %w", err)
	}
	var v int
	err := s.db.QueryRow(`SELECT version FROM centrai_schema_version LIMIT 1`).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		if _, err := s.db.Exec(`INSERT INTO centrai_schema_version (version) VALUES (?)`, schemaVersion); err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if v != schemaVersion {
		return fmt.Errorf("sqlite: unsupported schema version %d (want %d)", v, schemaVersion)
	}

	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS centrai_sessions (
		id TEXT PRIMARY KEY,
		data BLOB NOT NULL,
		updated_ms INTEGER NOT NULL
	)`)
	return err
}

// Close releases the database handle.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Load implements [session.Store].
func (s *Store) Load(ctx context.Context, id string) (*session.Session, error) {
	var blob []byte
	err := s.db.QueryRowContext(ctx, `SELECT data FROM centrai_sessions WHERE id = ?`, id).Scan(&blob)
	if errors.Is(err, sql.ErrNoRows) {
		return &session.Session{ID: id}, nil
	}
	if err != nil {
		return nil, err
	}
	var sess session.Session
	if err := json.Unmarshal(blob, &sess); err != nil {
		return nil, fmt.Errorf("sqlite: decode session: %w", err)
	}
	sess.ID = id
	return &sess, nil
}

// Save implements [session.Store].
func (s *Store) Save(ctx context.Context, sess *session.Session) error {
	if sess == nil {
		return nil
	}
	if strings.TrimSpace(sess.ID) == "" {
		return errors.New("sqlite: empty session id")
	}
	blob, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO centrai_sessions (id, data, updated_ms) VALUES (?, ?, strftime('%s','now')*1000)
		ON CONFLICT(id) DO UPDATE SET data = excluded.data, updated_ms = excluded.updated_ms`, sess.ID, blob)
	return err
}
