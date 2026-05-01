package recording

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"stargrazer/internal/automation"
	stardb "stargrazer/internal/db"
)

// SQLiteRepo persists Recording rows in the recordings table.
type SQLiteRepo struct {
	exec stardb.Executor
}

func NewSQLiteRepo(exec stardb.Executor) *SQLiteRepo { return &SQLiteRepo{exec: exec} }

func (r *SQLiteRepo) Save(rec *Recording) error {
	if rec.PlatformID == "" {
		return errors.New("save: platformID required")
	}
	if rec.Source == "" {
		rec.Source = "chrome-devtools-recorder"
	}
	if rec.ParsedSteps == nil {
		rec.ParsedSteps = []automation.Step{}
	}
	if rec.Warnings == nil {
		rec.Warnings = []string{}
	}
	now := time.Now().UTC()
	if rec.ID == "" {
		rec.ID = uuid.NewString()
		rec.CreatedAt = now
	}
	rec.UpdatedAt = now
	stepsJSON, err := json.Marshal(rec.ParsedSteps)
	if err != nil {
		return fmt.Errorf("marshal parsed_steps: %w", err)
	}
	warnJSON, err := json.Marshal(rec.Warnings)
	if err != nil {
		return fmt.Errorf("marshal warnings: %w", err)
	}
	_, err = r.exec.Exec(`
		INSERT INTO recordings(id, platform_id, title, source, raw_json, parsed_steps, warnings, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			platform_id  = excluded.platform_id,
			title        = excluded.title,
			source       = excluded.source,
			raw_json     = excluded.raw_json,
			parsed_steps = excluded.parsed_steps,
			warnings     = excluded.warnings,
			updated_at   = excluded.updated_at`,
		rec.ID, rec.PlatformID, rec.Title, rec.Source, rec.RawJSON,
		string(stepsJSON), string(warnJSON),
		rec.CreatedAt.Format(time.RFC3339Nano),
		rec.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("upsert recording %s: %w", rec.ID, err)
	}
	return nil
}

func (r *SQLiteRepo) Get(id string) (*Recording, error) {
	row := r.exec.QueryRow(`SELECT id, platform_id, title, source, raw_json, parsed_steps, warnings, created_at, updated_at FROM recordings WHERE id = ?`, id)
	rec, err := scanRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return rec, err
}

func (r *SQLiteRepo) List(platformID string) ([]Recording, error) {
	rows, err := r.exec.Query(`
		SELECT id, platform_id, title, source, raw_json, parsed_steps, warnings, created_at, updated_at
		FROM recordings WHERE platform_id = ? ORDER BY created_at DESC`, platformID)
	if err != nil {
		return nil, fmt.Errorf("query recordings: %w", err)
	}
	defer rows.Close()
	var out []Recording
	for rows.Next() {
		rec, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *rec)
	}
	return out, rows.Err()
}

func (r *SQLiteRepo) Delete(id string) error {
	res, err := r.exec.Exec(`DELETE FROM recordings WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete recording %s: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRow(row rowScanner) (*Recording, error) {
	var (
		rec           Recording
		stepsJSON     string
		warnJSON      string
		createdAt, ua string
	)
	if err := row.Scan(&rec.ID, &rec.PlatformID, &rec.Title, &rec.Source, &rec.RawJSON, &stepsJSON, &warnJSON, &createdAt, &ua); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(stepsJSON), &rec.ParsedSteps); err != nil {
		return nil, fmt.Errorf("unmarshal parsed_steps for %s: %w", rec.ID, err)
	}
	if err := json.Unmarshal([]byte(warnJSON), &rec.Warnings); err != nil {
		return nil, fmt.Errorf("unmarshal warnings for %s: %w", rec.ID, err)
	}
	t, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	rec.CreatedAt = t
	t, err = time.Parse(time.RFC3339Nano, ua)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	rec.UpdatedAt = t
	return &rec, nil
}

var _ Repository = (*SQLiteRepo)(nil)
