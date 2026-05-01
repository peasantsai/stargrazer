package social

// SessionRepo is the persistence seam for per-platform login state. The
// existing JSON-backed SessionStore satisfies a structurally compatible API
// but is retained only as a migrator read source (deleted in T7).
type SessionRepo interface {
	GetAll() []AccountStatus
	Get(id Platform) AccountStatus
	SetLoggedIn(id Platform, username string)
	SetLoggedOut(id Platform)
	UpdateCheckTime(id Platform)
}
