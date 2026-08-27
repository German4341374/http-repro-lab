BEGIN;
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE analysis_sessions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  source_sha256 char(64) NOT NULL,
  status text NOT NULL CHECK (status IN ('pending', 'running', 'complete', 'failed', 'cancelled')),
  created_at timestamptz NOT NULL DEFAULT now(),
  completed_at timestamptz,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);
CREATE TABLE sources (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  analysis_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
  kind text NOT NULL CHECK (kind IN ('har', 'curl', 'raw-http', 'postman', 'openapi', 'request-spec')),
  sha256 char(64) NOT NULL,
  byte_size bigint NOT NULL CHECK (byte_size >= 0),
  original_name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (analysis_id, sha256)
);
CREATE TABLE requests (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  analysis_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
  source_id uuid NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
  source_index integer NOT NULL CHECK (source_index >= 0),
  fingerprint char(64) NOT NULL,
  method text NOT NULL,
  scheme text NOT NULL CHECK (scheme IN ('http', 'https')),
  host text NOT NULL,
  port integer CHECK (port BETWEEN 1 AND 65535),
  path text NOT NULL,
  body_type text NOT NULL,
  body_sha256 char(64),
  normalized jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (source_id, source_index)
);
CREATE TABLE responses (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id uuid NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
  environment text NOT NULL,
  status_code integer CHECK (status_code BETWEEN 100 AND 599),
  content_type text,
  size_bytes bigint CHECK (size_bytes >= 0),
  body_sha256 char(64),
  truncated boolean NOT NULL DEFAULT false,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE request_headers (
  request_id uuid NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
  position integer NOT NULL CHECK (position >= 0),
  name text NOT NULL,
  value_placeholder text NOT NULL,
  is_sensitive boolean NOT NULL DEFAULT false,
  PRIMARY KEY (request_id, position)
);
CREATE TABLE response_headers (
  response_id uuid NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
  position integer NOT NULL CHECK (position >= 0),
  name text NOT NULL,
  value text NOT NULL,
  PRIMARY KEY (response_id, position)
);
CREATE TABLE timing_events (
  response_id uuid NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
  phase text NOT NULL CHECK (phase IN ('blocked', 'dns', 'connect', 'tls', 'send', 'ttfb', 'receive', 'total')),
  duration_ms numeric(14,3) NOT NULL CHECK (duration_ms >= 0),
  PRIMARY KEY (response_id, phase)
);
CREATE TABLE findings (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  analysis_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
  request_id uuid REFERENCES requests(id) ON DELETE CASCADE,
  rule_id text NOT NULL,
  severity text NOT NULL CHECK (severity IN ('info', 'low', 'medium', 'high', 'critical')),
  category text NOT NULL,
  title text NOT NULL,
  summary text NOT NULL,
  confidence text NOT NULL CHECK (confidence IN ('exact', 'strong', 'heuristic')),
  evidence jsonb NOT NULL,
  next_steps jsonb NOT NULL DEFAULT '[]'::jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE sensitive_values (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  analysis_id uuid NOT NULL REFERENCES analysis_sessions(id) ON DELETE CASCADE,
  request_id uuid REFERENCES requests(id) ON DELETE CASCADE,
  kind text NOT NULL,
  location text NOT NULL,
  confidence text NOT NULL,
  preview text NOT NULL,
  replacement text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE reproductions (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id uuid NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
  target_host text NOT NULL,
  status text NOT NULL CHECK (status IN ('queued', 'running', 'complete', 'failed', 'blocked', 'cancelled')),
  error_code text,
  response_id uuid REFERENCES responses(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  completed_at timestamptz
);
CREATE TABLE comparisons (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id uuid NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
  response_a_id uuid NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
  response_b_id uuid NOT NULL REFERENCES responses(id) ON DELETE CASCADE,
  differences jsonb NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE repro_packs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id uuid NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
  manifest_sha256 char(64) NOT NULL,
  sanitization_profile text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE generated_clients (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  request_id uuid NOT NULL REFERENCES requests(id) ON DELETE CASCADE,
  language text NOT NULL,
  generator_version text NOT NULL,
  source_sha256 char(64) NOT NULL,
  verified_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (request_id, language, generator_version)
);
CREATE TABLE audit_events (
  id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  analysis_id uuid REFERENCES analysis_sessions(id) ON DELETE SET NULL,
  event text NOT NULL,
  actor text NOT NULL DEFAULT 'local-user',
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_created_at ON analysis_sessions (created_at DESC);
CREATE INDEX idx_sessions_status ON analysis_sessions (status, created_at DESC);
CREATE INDEX idx_requests_analysis ON requests (analysis_id, source_index);
CREATE INDEX idx_requests_fingerprint ON requests (fingerprint);
CREATE INDEX idx_requests_host_path ON requests (host, path);
CREATE INDEX idx_findings_analysis_severity ON findings (analysis_id, severity);
CREATE INDEX idx_findings_rule ON findings (rule_id, created_at DESC);
CREATE INDEX idx_comparisons_request_created ON comparisons (request_id, created_at DESC);
CREATE INDEX idx_audit_analysis_time ON audit_events (analysis_id, occurred_at DESC);
COMMIT;

