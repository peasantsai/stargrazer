package scheduler

// ScheduleRepo is the persistence seam for cron-managed jobs. The runtime
// cron handle (cronEntryID) is NOT persisted; it's regenerated when the
// scheduler registers jobs at startup.
type ScheduleRepo interface {
	List() ([]*Job, error)
	Get(id string) (*Job, error)
	Save(j *Job) error
	Delete(id string) error
}
