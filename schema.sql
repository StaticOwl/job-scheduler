-- Create the database
CREATE DATABASE scheduler_db;

-- Connect to scheduler_db and run the rest
\c scheduler_db;

-- Create the jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    command TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'queued',
    last_run TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on status for faster queries
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);

-- Insert some sample jobs
INSERT INTO jobs (name, command, status) VALUES
    ('test_job_1', 'echo "Job 1 is running"', 'queued'),
    ('test_job_2', 'echo "Job 2 is running"', 'queued')
ON CONFLICT (name) DO NOTHING;
