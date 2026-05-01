package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Migrate applies every embedded migration whose version is not yet recorded
// in schema_migrations, in lexicographic filename order. Each migration runs
// in its own transaction.
func Migrate(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return err
	}
	applied, err := appliedVersionsSet(db)
	if err != nil {
		return err
	}
	files, err := listMigrations()
	if err != nil {
		return err
	}
	for _, f := range files {
		version := strings.TrimSuffix(f, ".sql")
		if applied[version] {
			continue
		}
		body, err := migrationFS.ReadFile("migrations/" + f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		err = WithTx(db, func(tx *sql.Tx) error {
			if _, err := tx.Exec(string(body)); err != nil {
				return fmt.Errorf("apply %s: %w", f, err)
			}
			if _, err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, version); err != nil {
				return fmt.Errorf("record %s: %w", f, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// AppliedVersions returns the sorted list of applied migration versions.
func AppliedVersions(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		return nil, fmt.Errorf("query schema_migrations: %w", err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}
	return nil
}

func appliedVersionsSet(db *sql.DB) (map[string]bool, error) {
	versions, err := AppliedVersions(db)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(versions))
	for _, v := range versions {
		out[v] = true
	}
	return out, nil
}

func listMigrations() ([]string, error) {
	entries, err := fs.ReadDir(migrationFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() && strings.HasSuffix(name, ".sql") {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out, nil
}
