// Package profile persists named variable maps used by the planner to fill
// {{var}} placeholders at run time. Production read/write goes through
// SQLiteRepo (sqlite_repo.go), which satisfies Repository.
package profile

import (
	"errors"
	"time"
)

// ErrNotFound is returned when Get/GetByName/Delete cannot find the requested row.
var ErrNotFound = errors.New("profile: not found")

// Profile is a named variable bag used by the planner.
type Profile struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Vars      map[string]any `json:"vars"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// Repository is the persistence seam for variable profiles.
type Repository interface {
	Save(p *Profile) error
	Get(id string) (*Profile, error)
	GetByName(name string) (*Profile, error)
	List() ([]Profile, error)
	Delete(id string) error
}
