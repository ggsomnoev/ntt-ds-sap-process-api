BEGIN;

CREATE TABLE IF NOT EXISTS process_definitions (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS process_events (    
    uuid UUID PRIMARY KEY,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE process_runs (
    id UUID PRIMARY KEY,
    definition JSONB NOT NULL,
    status TEXT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ
);

CREATE TABLE process_logs (
    id SERIAL PRIMARY KEY,
    process_id UUID NOT NULL REFERENCES process_runs(id) ON DELETE CASCADE,
    log TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);


COMMIT;