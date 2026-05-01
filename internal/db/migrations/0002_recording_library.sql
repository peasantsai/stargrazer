-- 0002_recording_library.sql — DSG-001-P3 recordings, step templates, variable profiles.

CREATE TABLE recordings (
  id            TEXT PRIMARY KEY,
  platform_id   TEXT NOT NULL,
  title         TEXT NOT NULL,
  source        TEXT NOT NULL,
  raw_json      TEXT NOT NULL,
  parsed_steps  TEXT NOT NULL,
  warnings      TEXT NOT NULL,
  created_at    DATETIME NOT NULL,
  updated_at    DATETIME NOT NULL
);
CREATE INDEX idx_recordings_platform ON recordings(platform_id);

CREATE TABLE step_templates (
  id            TEXT PRIMARY KEY,
  name          TEXT NOT NULL,
  description   TEXT NOT NULL DEFAULT '',
  platform_id   TEXT,
  steps_json    TEXT NOT NULL,
  required_vars TEXT NOT NULL,
  created_at    DATETIME NOT NULL,
  updated_at    DATETIME NOT NULL,
  UNIQUE(platform_id, name)
);
CREATE INDEX idx_step_templates_platform ON step_templates(platform_id);

CREATE TABLE variable_profiles (
  id          TEXT PRIMARY KEY,
  name        TEXT NOT NULL UNIQUE,
  vars_json   TEXT NOT NULL,
  created_at  DATETIME NOT NULL,
  updated_at  DATETIME NOT NULL
);

ALTER TABLE automations ADD COLUMN default_profile_id TEXT;
ALTER TABLE schedules   ADD COLUMN profile_id TEXT;
