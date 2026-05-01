package automation

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	stardb "stargrazer/internal/db"
)

// SQLiteRepo persists Config rows in the `automations` table. The exec
// field can be either a *sql.DB (production) or a *sql.Tx (migrator).
type SQLiteRepo struct {
	exec stardb.Executor
}

// NewSQLiteRepo constructs a SQLiteRepo bound to exec. Caller owns lifecycle.
func NewSQLiteRepo(exec stardb.Executor) *SQLiteRepo { return &SQLiteRepo{exec: exec} }

func (r *SQLiteRepo) List(platformID string) ([]Config, error) {
	rows, err := r.exec.Query(`
		SELECT id, platform_id, name, description, steps, created_at, last_run, run_count
		FROM automations
		WHERE platform_id = ?
		ORDER BY created_at`, platformID)
	if err != nil {
		return nil, fmt.Errorf("query automations: %w", err)
	}
	defer rows.Close()

	var out []Config
	for rows.Next() {
		var (
			cfg       Config
			stepsJSON string
			createdAt string
			lastRun   sql.NullString
		)
		if err := rows.Scan(&cfg.ID, &cfg.PlatformID, &cfg.Name, &cfg.Description, &stepsJSON, &createdAt, &lastRun, &cfg.RunCount); err != nil {
			return nil, fmt.Errorf("scan automation row: %w", err)
		}
		if err := json.Unmarshal([]byte(stepsJSON), &cfg.Steps); err != nil {
			return nil, fmt.Errorf("unmarshal steps for %s: %w", cfg.ID, err)
		}
		t, err := time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse created_at for %s: %w", cfg.ID, err)
		}
		cfg.CreatedAt = t
		cfg.LastRun = stardb.ParseNullTime(lastRun)
		out = append(out, cfg)
	}
	return out, rows.Err()
}

func (r *SQLiteRepo) Save(cfg Config) (Config, error) {
	if cfg.PlatformID == "" {
		return Config{}, errors.New("Save: PlatformID required")
	}
	if cfg.ID == "" {
		cfg.ID = uuid.NewString()
		cfg.CreatedAt = time.Now()
	}
	if cfg.Steps == nil {
		cfg.Steps = []Step{}
	}
	stepsJSON, err := json.Marshal(cfg.Steps)
	if err != nil {
		return Config{}, fmt.Errorf("marshal steps: %w", err)
	}
	_, err = r.exec.Exec(`
		INSERT INTO automations(id, platform_id, name, description, steps, created_at, last_run, run_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			platform_id = excluded.platform_id,
			name        = excluded.name,
			description = excluded.description,
			steps       = excluded.steps,
			last_run    = excluded.last_run,
			run_count   = excluded.run_count`,
		cfg.ID, cfg.PlatformID, cfg.Name, cfg.Description, string(stepsJSON),
		cfg.CreatedAt.UTC().Format(time.RFC3339Nano),
		stardb.FormatNullTime(cfg.LastRun),
		cfg.RunCount,
	)
	if err != nil {
		return Config{}, fmt.Errorf("upsert automation %s: %w", cfg.ID, err)
	}
	return cfg, nil
}

func (r *SQLiteRepo) Delete(platformID, id string) (bool, error) {
	res, err := r.exec.Exec(`DELETE FROM automations WHERE platform_id = ? AND id = ?`, platformID, id)
	if err != nil {
		return false, fmt.Errorf("delete automation %s: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *SQLiteRepo) RecordRun(platformID, id string) error {
	res, err := r.exec.Exec(`
		UPDATE automations
		SET last_run = ?, run_count = run_count + 1
		WHERE platform_id = ? AND id = ?`,
		time.Now().UTC().Format(time.RFC3339Nano), platformID, id)
	if err != nil {
		return fmt.Errorf("record run for %s: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("automation %q not found for platform %q", id, platformID)
	}
	return nil
}

// Compile-time assertion: *SQLiteRepo satisfies Repository.
var _ Repository = (*SQLiteRepo)(nil)
