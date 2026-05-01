package scheduler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	stardb "stargrazer/internal/db"
)

// SQLiteRepo persists Job rows in the schedules table. The exec field can
// be either a *sql.DB (production) or a *sql.Tx (migrator).
type SQLiteRepo struct {
	exec stardb.Executor
}

// NewSQLiteRepo constructs a SQLiteRepo bound to exec.
func NewSQLiteRepo(exec stardb.Executor) *SQLiteRepo { return &SQLiteRepo{exec: exec} }

func (r *SQLiteRepo) List() ([]*Job, error) {
	rows, err := r.exec.Query(`
		SELECT id, name, type, platforms, cron_expr, status, next_run, last_run,
		       run_count, last_result, auto, upload_config, created_at
		FROM schedules
		ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("query schedules: %w", err)
	}
	defer rows.Close()

	var out []*Job
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *SQLiteRepo) Get(id string) (*Job, error) {
	row := r.exec.QueryRow(`
		SELECT id, name, type, platforms, cron_expr, status, next_run, last_run,
		       run_count, last_result, auto, upload_config, created_at
		FROM schedules
		WHERE id = ?`, id)
	j, err := scanJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("schedule %q: %w", id, sql.ErrNoRows)
	}
	return j, err
}

func (r *SQLiteRepo) Save(j *Job) error {
	if j == nil {
		return errors.New("save: nil job")
	}
	if j.ID == "" {
		return errors.New("save: ID required (caller assigns)")
	}
	platformsJSON, err := json.Marshal(j.Platforms)
	if err != nil {
		return fmt.Errorf("marshal platforms: %w", err)
	}
	var uploadJSON sql.NullString
	if j.UploadConfig != nil {
		b, err := json.Marshal(j.UploadConfig)
		if err != nil {
			return fmt.Errorf("marshal upload config: %w", err)
		}
		uploadJSON = sql.NullString{String: string(b), Valid: true}
	}
	autoInt := 0
	if j.Auto {
		autoInt = 1
	}
	_, err = r.exec.Exec(`
		INSERT INTO schedules(id, name, type, platforms, cron_expr, status,
		                     next_run, last_run, run_count, last_result, auto,
		                     upload_config, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name          = excluded.name,
			type          = excluded.type,
			platforms     = excluded.platforms,
			cron_expr     = excluded.cron_expr,
			status        = excluded.status,
			next_run      = excluded.next_run,
			last_run      = excluded.last_run,
			run_count     = excluded.run_count,
			last_result   = excluded.last_result,
			auto          = excluded.auto,
			upload_config = excluded.upload_config`,
		j.ID, j.Name, string(j.Type), string(platformsJSON), j.CronExpr, string(j.Status),
		stardb.FormatNullTime(j.NextRun),
		stardb.FormatNullTime(j.LastRun),
		j.RunCount, j.LastResult, autoInt,
		uploadJSON,
		j.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("upsert schedule %s: %w", j.ID, err)
	}
	return nil
}

func (r *SQLiteRepo) Delete(id string) error {
	res, err := r.exec.Exec(`DELETE FROM schedules WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete schedule %s: %w", id, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("schedule %q not found", id)
	}
	return nil
}

// rowScanner is satisfied by both *sql.Row and *sql.Rows.
type rowScanner interface {
	Scan(dest ...any) error
}

func scanJob(s rowScanner) (*Job, error) {
	var (
		j            Job
		typ          string
		status       string
		platformsRaw string
		nextRun      sql.NullString
		lastRun      sql.NullString
		autoInt      int
		uploadConfig sql.NullString
		createdAt    string
	)
	if err := s.Scan(&j.ID, &j.Name, &typ, &platformsRaw, &j.CronExpr, &status,
		&nextRun, &lastRun, &j.RunCount, &j.LastResult, &autoInt, &uploadConfig, &createdAt); err != nil {
		return nil, err
	}
	j.Type = JobType(typ)
	j.Status = JobStatus(status)
	j.Auto = autoInt != 0
	if err := json.Unmarshal([]byte(platformsRaw), &j.Platforms); err != nil {
		return nil, fmt.Errorf("unmarshal platforms for %s: %w", j.ID, err)
	}
	if uploadConfig.Valid && uploadConfig.String != "" {
		var uc UploadConfig
		if err := json.Unmarshal([]byte(uploadConfig.String), &uc); err != nil {
			return nil, fmt.Errorf("unmarshal upload_config for %s: %w", j.ID, err)
		}
		j.UploadConfig = &uc
	}
	j.NextRun = stardb.ParseNullTime(nextRun)
	j.LastRun = stardb.ParseNullTime(lastRun)
	t, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at for %s: %w", j.ID, err)
	}
	j.CreatedAt = t
	return &j, nil
}

// Compile-time assertion: *SQLiteRepo satisfies ScheduleRepo.
var _ ScheduleRepo = (*SQLiteRepo)(nil)
