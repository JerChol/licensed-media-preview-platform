CREATE TABLE IF NOT EXISTS sources (
  source_id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  license_type TEXT NOT NULL,
  attribution_text TEXT NOT NULL,
  match_hosts TEXT[] NOT NULL,
  path_prefixes TEXT[] NOT NULL,
  file_extensions TEXT[] NOT NULL,
  allowed_media_kind TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS jobs (
  id UUID PRIMARY KEY,
  source_url TEXT NOT NULL,
  resolved_url TEXT NOT NULL,
  resolved_source_id TEXT REFERENCES sources(source_id),
  status TEXT NOT NULL DEFAULT 'queued',
  progress INTEGER NOT NULL DEFAULT 0,
  error_message TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS artifacts (
  id UUID PRIMARY KEY,
  job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
  artifact_tye TEXT NOT NULL,
  storage_path TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);