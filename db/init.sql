CREATE TABLE jobs (
    id              TEXT PRIMARY KEY,
    idempotency_key TEXT,
    type            TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL,
    priority        INT NOT NULL DEFAULT 0,
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 5,
    run_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_jobs_idem ON jobs (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX idx_jobs_status ON jobs (status, run_at);