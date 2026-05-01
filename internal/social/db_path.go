package social

import "path/filepath"

// DBPath returns the absolute filesystem path of the Stargrazer SQLite database.
// Lives next to the existing sessions/ directory: <base>/stargrazer.db.
func DBPath() string {
	return filepath.Join(sessionsBaseDir(), "stargrazer.db")
}
