package automation

// Repository is the persistence seam consumed by App. Concrete implementations
// (SQLite in production, JSON-legacy in the migrator) satisfy this interface.
type Repository interface {
	List(platformID string) ([]Config, error)
	Save(cfg Config) (Config, error)
	Delete(platformID, id string) (bool, error)
	RecordRun(platformID, id string) error
}
