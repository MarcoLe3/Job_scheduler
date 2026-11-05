CREATE TABLE jobs (
    id              TEXT PRIMARY KEY,
    idempotency_key TEXT,
    type            TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL,
    priority        INT NOT NULL DEFAULT 0,
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 5,
    
)