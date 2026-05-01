package db

import (
	"database/sql"
	"time"
)

// FormatNullTime returns a value suitable for sql.DB.Exec when a column may
// be NULL. Zero times become NULL; non-zero times become an RFC3339Nano string
// in UTC.
func FormatNullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}

// ParseNullTime parses a sql.NullString back to time.Time. Invalid or empty
// strings return the zero value, never an error.
func ParseNullTime(s sql.NullString) time.Time {
	if !s.Valid || s.String == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s.String)
	if err != nil {
		return time.Time{}
	}
	return t
}
