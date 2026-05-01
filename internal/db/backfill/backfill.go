// Package backfill performs the one-shot JSON → SQLite migration on first
// SQLite-enabled launch. Idempotent: subsequent calls return immediately if
// the sentinel row in schema_migrations is present.
package backfill

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"stargrazer/internal/automation"
	stardb "stargrazer/internal/db"
	"stargrazer/internal/scheduler"
	"stargrazer/internal/social"
)

const sentinel = "backfill_p2"

// RunIfNeeded checks the sentinel; if absent, reads JSON sources from baseDir,
// writes them into db via the SQLite repos in one transaction, sets the
// sentinel, and renames source files to <original>.preP2.bak.
func RunIfNeeded(db *sql.DB, baseDir string) error {
	already, err := sentinelExists(db)
	if err != nil {
		return err
	}
	if already {
		return nil
	}

	autoDir := filepath.Join(baseDir, "automations")
	accountsPath := filepath.Join(baseDir, "accounts.json")
	schedulesPath := filepath.Join(baseDir, "schedules.json")

	autoFiles, err := readAutomations(autoDir)
	if err != nil {
		return fmt.Errorf("read automations: %w", err)
	}
	accounts, err := readAccounts(accountsPath)
	if err != nil {
		return fmt.Errorf("read accounts: %w", err)
	}
	jobs, err := readSchedules(schedulesPath)
	if err != nil {
		return fmt.Errorf("read schedules: %w", err)
	}

	// All writes happen inside one transaction. Repos accept stardb.Executor
	// (satisfied by both *sql.DB and *sql.Tx) so we construct them with tx
	// inside WithTx and the writes participate in the transaction.
	err = stardb.WithTx(db, func(tx *sql.Tx) error {
		autoRepo := automation.NewSQLiteRepo(tx)
		schedRepo := scheduler.NewSQLiteRepo(tx)

		for _, cfg := range autoFiles {
			if _, err := autoRepo.Save(cfg); err != nil {
				return fmt.Errorf("backfill automation %s: %w", cfg.ID, err)
			}
		}

		// Accounts: write directly via tx so DB errors abort the migration.
		// SessionRepo's Set* methods log-and-swallow at runtime to match the
		// legacy JSON contract, which is the wrong shape for the migrator.
		for _, acc := range accounts {
			loggedIn := 0
			if acc.LoggedIn {
				loggedIn = 1
			}
			_, err := tx.Exec(`
				INSERT INTO accounts(id, username, logged_in, last_login, last_check)
				VALUES (?, ?, ?, ?, ?)
				ON CONFLICT(id) DO UPDATE SET
					username   = excluded.username,
					logged_in  = excluded.logged_in,
					last_login = excluded.last_login,
					last_check = excluded.last_check`,
				string(acc.PlatformID), acc.Username, loggedIn,
				stardb.FormatNullTime(acc.LastLogin),
				stardb.FormatNullTime(acc.LastCheck),
			)
			if err != nil {
				return fmt.Errorf("backfill account %s: %w", acc.PlatformID, err)
			}
		}

		for _, j := range jobs {
			if err := schedRepo.Save(j); err != nil {
				return fmt.Errorf("backfill schedule %s: %w", j.ID, err)
			}
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES (?)`, sentinel); err != nil {
			return fmt.Errorf("write sentinel: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Archive sources. Failures here are warnings.
	for _, p := range archivePaths(autoDir, accountsPath, schedulesPath) {
		if err := archive(p); err != nil {
			fmt.Fprintf(os.Stderr, "backfill: archive %s: %v\n", p, err)
		}
	}

	fmt.Printf("backfill: %d accounts, %d schedules, %d automations migrated\n",
		len(accounts), len(jobs), len(autoFiles))
	return nil
}

func sentinelExists(db *sql.DB) (bool, error) {
	var v string
	err := db.QueryRow(`SELECT version FROM schema_migrations WHERE version = ?`, sentinel).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check sentinel: %w", err)
	}
	return v == sentinel, nil
}

func readAutomations(dir string) ([]automation.Config, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var out []automation.Config
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		body, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", e.Name(), err)
		}
		var cfgs []automation.Config
		if err := json.Unmarshal(body, &cfgs); err != nil {
			return nil, fmt.Errorf("parse %s: %w", e.Name(), err)
		}
		out = append(out, cfgs...)
	}
	return out, nil
}

func readAccounts(path string) ([]social.AccountStatus, error) {
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var accounts []social.AccountStatus
	if err := json.Unmarshal(body, &accounts); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return accounts, nil
}

func readSchedules(path string) ([]*scheduler.Job, error) {
	body, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var jobs []*scheduler.Job
	if err := json.Unmarshal(body, &jobs); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	for _, j := range jobs {
		if j.CreatedAt.IsZero() {
			j.CreatedAt = time.Now()
		}
	}
	return jobs, nil
}
