package profile

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	stardb "stargrazer/internal/db"
)

// SQLiteRepo persists Profile rows in the variable_profiles table.
type SQLiteRepo struct {
	exec stardb.Executor
}

func NewSQLiteRepo(exec stardb.Executor) *SQLiteRepo { return &SQLiteRepo{exec: exec} }

func (r *SQLiteRepo) Save(p *Profile) error {
	if p.Name == "" {
		return errors.New("save: name required")
	}
	if p.Vars == nil {
		p.Vars = map[string]any{}
	}
	now := time.Now().UTC()
	if p.ID == "" {
		p.ID = uuid.NewString()
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	varsJSON, err := json.Marshal(p.Vars)
	if err != nil {
		return fmt.Errorf("marshal vars: %w", err)
	}
	_, err = r.exec.Exec(`
		INSERT INTO variable_profiles(id, name, vars_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name       = excluded.name,
			vars_json  = excluded.vars_json,
			updated_at = excluded.updated_at`,
		p.ID, p.Name, string(varsJSON),
		p.CreatedAt.Format(time.RFC3339Nano),
		p.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("upsert profile %s: %w", p.ID, err)
	}
	return nil
}

func (r *SQLiteRepo) Get(id string) (*Profile, error) {
	row := r.exec.QueryRow(`SELECT id, name, vars_json, created_at, updated_at FROM variable_profiles WHERE id = ?`, id)
	return scanOne(row)
}

func (r *SQLiteRepo) GetByName(name string) (*Profile, error) {
	row := r.exec.QueryRow(`SELECT id, name, vars_json, created_at, updated_at FROM variable_profiles WHERE name = ?`, name)
	return scanOne(row)
}

func (r *SQLiteRepo) List() ([]Profile, error) {
	rows, err := r.exec.Query(`SELECT id, name, vars_json, created_at, updated_at FROM variable_profiles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query profiles: %w", err)
	}
	defer rows.Close()
	var out []Profile
	for rows.Next() {
		p, err := scanRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *p)
	}
	return out, rows.Err()
}

func (r *SQLiteRepo) Delete(id string) error {
	res, err := r.exec.Exec(`DELETE FROM variable_profiles WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete profile %s: %w", id, err)
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

func scanOne(row rowScanner) (*Profile, error) {
	p, err := scanRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func scanRow(row rowScanner) (*Profile, error) {
	var (
		p             Profile
		varsJSON      string
		createdAt, ua string
	)
	if err := row.Scan(&p.ID, &p.Name, &varsJSON, &createdAt, &ua); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(varsJSON), &p.Vars); err != nil {
		return nil, fmt.Errorf("unmarshal vars for %s: %w", p.ID, err)
	}
	t, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at for %s: %w", p.ID, err)
	}
	p.CreatedAt = t
	t, err = time.Parse(time.RFC3339Nano, ua)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at for %s: %w", p.ID, err)
	}
	p.UpdatedAt = t
	return &p, nil
}

var _ Repository = (*SQLiteRepo)(nil)
