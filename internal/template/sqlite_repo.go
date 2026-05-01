package template

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

// SQLiteRepo persists Template rows in the step_templates table.
type SQLiteRepo struct {
	exec stardb.Executor
}

func NewSQLiteRepo(exec stardb.Executor) *SQLiteRepo { return &SQLiteRepo{exec: exec} }

func (r *SQLiteRepo) Save(t *Template) error {
	if t.Name == "" {
		return errors.New("save: name required")
	}
	for i, s := range t.Steps {
		if s.Action == automation.ActionTemplate {
			return fmt.Errorf("step %d: %w", i, ErrNestedTemplate)
		}
	}
	if t.Steps == nil {
		t.Steps = []automation.Step{}
	}
	if t.RequiredVars == nil {
		t.RequiredVars = []string{}
	}
	now := time.Now().UTC()
	if t.ID == "" {
		t.ID = uuid.NewString()
		t.CreatedAt = now
	}
	t.UpdatedAt = now
	stepsJSON, err := json.Marshal(t.Steps)
	if err != nil {
		return fmt.Errorf("marshal steps: %w", err)
	}
	reqJSON, err := json.Marshal(t.RequiredVars)
	if err != nil {
		return fmt.Errorf("marshal required_vars: %w", err)
	}
	_, err = r.exec.Exec(`
		INSERT INTO step_templates(id, name, description, platform_id, steps_json, required_vars, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name          = excluded.name,
			description   = excluded.description,
			platform_id   = excluded.platform_id,
			steps_json    = excluded.steps_json,
			required_vars = excluded.required_vars,
			updated_at    = excluded.updated_at`,
		t.ID, t.Name, t.Description, nullableString(t.PlatformID),
		string(stepsJSON), string(reqJSON),
		t.CreatedAt.Format(time.RFC3339Nano),
		t.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("upsert template %s: %w", t.ID, err)
	}
	return nil
}

func (r *SQLiteRepo) Get(id string) (*Template, error) {
	row := r.exec.QueryRow(`SELECT id, name, description, platform_id, steps_json, required_vars, created_at, updated_at FROM step_templates WHERE id = ?`, id)
	t, err := scanRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func (r *SQLiteRepo) GetByName(platformID *string, name string) (*Template, error) {
	var row *sql.Row
	if platformID == nil {
		row = r.exec.QueryRow(`SELECT id, name, description, platform_id, steps_json, required_vars, created_at, updated_at FROM step_templates WHERE platform_id IS NULL AND name = ?`, name)
	} else {
		row = r.exec.QueryRow(`SELECT id, name, description, platform_id, steps_json, required_vars, created_at, updated_at FROM step_templates WHERE platform_id = ? AND name = ?`, *platformID, name)
	}
	t, err := scanRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func (r *SQLiteRepo) List(platformID string) ([]Template, error) {
	rows, err := r.exec.Query(`
		SELECT id, name, description, platform_id, steps_json, required_vars, created_at, updated_at
		FROM step_templates
		WHERE platform_id IS NULL OR platform_id = ?
		ORDER BY name`, platformID)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()
	var out []Template
	for rows.Next() {
		t, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}

func (r *SQLiteRepo) Delete(id string) error {
	res, err := r.exec.Exec(`DELETE FROM step_templates WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete template %s: %w", id, err)
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

func scanRow(row rowScanner) (*Template, error) {
	var (
		t             Template
		platformID    sql.NullString
		stepsJSON     string
		requiredVars  string
		createdAt, ua string
	)
	if err := row.Scan(&t.ID, &t.Name, &t.Description, &platformID, &stepsJSON, &requiredVars, &createdAt, &ua); err != nil {
		return nil, err
	}
	if platformID.Valid {
		s := platformID.String
		t.PlatformID = &s
	}
	if err := json.Unmarshal([]byte(stepsJSON), &t.Steps); err != nil {
		return nil, fmt.Errorf("unmarshal steps for %s: %w", t.ID, err)
	}
	if err := json.Unmarshal([]byte(requiredVars), &t.RequiredVars); err != nil {
		return nil, fmt.Errorf("unmarshal required_vars for %s: %w", t.ID, err)
	}
	tt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}
	t.CreatedAt = tt
	tt, err = time.Parse(time.RFC3339Nano, ua)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}
	t.UpdatedAt = tt
	return &t, nil
}

func nullableString(p *string) any {
	if p == nil {
		return nil
	}
	return *p
}

var _ Repository = (*SQLiteRepo)(nil)
