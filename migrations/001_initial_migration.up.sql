BEGIN;

CREATE TABLE IF NOT EXISTS processed_files (    
    id SERIAL PRIMARY KEY,
    filename TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

COMMIT;