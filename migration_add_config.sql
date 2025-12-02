-- Migration to add config table for dynamic settings
\c scheduler_db;

-- Create config table
CREATE TABLE IF NOT EXISTS scheduler_config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default max_concurrent_jobs config
INSERT INTO scheduler_config (key, value) VALUES
    ('max_concurrent_jobs', '2')
ON CONFLICT (key) DO NOTHING;

-- Create index on key for faster lookups
CREATE INDEX IF NOT EXISTS idx_config_key ON scheduler_config(key);
