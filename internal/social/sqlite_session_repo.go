package social

import (
	"database/sql"
	"log"
	"time"

	stardb "stargrazer/internal/db"
)

// SQLiteSessionRepo persists per-platform AccountStatus in the accounts table.
// The exec field can be either a *sql.DB (production) or a *sql.Tx (migrator).
type SQLiteSessionRepo struct {
	exec stardb.Executor
}

// NewSQLiteSessionRepo constructs a SQLiteSessionRepo bound to exec.
func NewSQLiteSessionRepo(exec stardb.Executor) *SQLiteSessionRepo {
	return &SQLiteSessionRepo{exec: exec}
}

func (r *SQLiteSessionRepo) GetAll() []AccountStatus {
	platforms := AllPlatforms()
	out := make([]AccountStatus, 0, len(platforms))
	for _, p := range platforms {
		out = append(out, r.Get(p.ID))
	}
	return out
}

func (r *SQLiteSessionRepo) Get(id Platform) AccountStatus {
	var (
		username  string
		loggedIn  int
		lastLogin sql.NullString
		lastCheck sql.NullString
	)
	row := r.exec.QueryRow(
		`SELECT username, logged_in, last_login, last_check FROM accounts WHERE id = ?`,
		string(id),
	)
	err := row.Scan(&username, &loggedIn, &lastLogin, &lastCheck)
	if err == sql.ErrNoRows {
		return AccountStatus{PlatformID: id}
	}
	if err != nil {
		log.Printf("session_repo: Get(%s): %v", id, err)
		return AccountStatus{PlatformID: id}
	}
	return AccountStatus{
		PlatformID: id,
		LoggedIn:   loggedIn != 0,
		Username:   username,
		LastLogin:  stardb.ParseNullTime(lastLogin),
		LastCheck:  stardb.ParseNullTime(lastCheck),
	}
}

func (r *SQLiteSessionRepo) SetLoggedIn(id Platform, username string) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := r.exec.Exec(`
		INSERT INTO accounts(id, username, logged_in, last_login, last_check)
		VALUES (?, ?, 1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			username   = excluded.username,
			logged_in  = 1,
			last_login = excluded.last_login,
			last_check = excluded.last_check`,
		string(id), username, now, now)
	if err != nil {
		log.Printf("session_repo: SetLoggedIn(%s): %v", id, err)
	}
}

func (r *SQLiteSessionRepo) SetLoggedOut(id Platform) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := r.exec.Exec(`
		INSERT INTO accounts(id, username, logged_in, last_login, last_check)
		VALUES (?, '', 0, NULL, ?)
		ON CONFLICT(id) DO UPDATE SET
			username   = '',
			logged_in  = 0,
			last_check = excluded.last_check`,
		string(id), now)
	if err != nil {
		log.Printf("session_repo: SetLoggedOut(%s): %v", id, err)
	}
}

func (r *SQLiteSessionRepo) UpdateCheckTime(id Platform) {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := r.exec.Exec(`
		UPDATE accounts SET last_check = ? WHERE id = ?`,
		now, string(id))
	if err != nil {
		log.Printf("session_repo: UpdateCheckTime(%s): %v", id, err)
	}
}

// Compile-time assertion: *SQLiteSessionRepo satisfies SessionRepo.
var _ SessionRepo = (*SQLiteSessionRepo)(nil)
