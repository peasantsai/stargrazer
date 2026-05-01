-- 0001_init.sql — DSG-001-P2 initial schema.
-- Active domains migrated from JSON. Phase-3/5 tables ship with their phases.

CREATE TABLE IF NOT EXISTS schema_migrations (
  version    TEXT PRIMARY KEY,
  applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE accounts (
  id            TEXT PRIMARY KEY,
  username      TEXT NOT NULL DEFAULT '',
  logged_in     INTEGER NOT NULL DEFAULT 0,
  last_login    DATETIME,
  last_check    DATETIME
);

CREATE TABLE schedules (
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  type          TEXT NOT NULL,
  platforms     TEXT NOT NULL,
  cron_expr     TEXT NOT NULL,
  status        TEXT NOT NULL,
  next_run      DATETIME,
  last_run      DATETIME,
  run_count     INTEGER NOT NULL DEFAULT 0,
  last_result   TEXT NOT NULL DEFAULT '',
  auto          INTEGER NOT NULL DEFAULT 0,
  upload_config TEXT,
  created_at    DATETIME NOT NULL
);
CREATE INDEX idx_schedules_status ON schedules(status);

CREATE TABLE automations (
  id            TEXT PRIMARY KEY,
  platform_id   TEXT NOT NULL,
  name          TEXT NOT NULL,
  description   TEXT NOT NULL DEFAULT '',
  steps         TEXT NOT NULL,
  created_at    DATETIME NOT NULL,
  last_run      DATETIME,
  run_count     INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX idx_automations_platform ON automations(platform_id);
