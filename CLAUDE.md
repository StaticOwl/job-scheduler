# Job Scheduler Daemon

## Project Overview
A lightweight job scheduler daemon built in Go that monitors jobs in a PostgreSQL database and triggers them with concurrent execution limits.

## Tech Stack
- **Language**: Go 1.25.4
- **Database**: PostgreSQL 17.5
- **Platform**: Windows

## Features
1. **Web Dashboard** - Real-time job visualization UI with auto-refresh (http://localhost:8080)
2. **REST API** - Full HTTP API for job and config management
3. **Continuous Daemon** - Runs continuously, checking for jobs at regular intervals
4. **Queued Job Execution** - Picks up jobs with "queued" status and executes them
5. **Concurrent Job Limiting** - Limits number of jobs running simultaneously
6. **Dynamic Config Reload** - Changes to max_concurrent_jobs take effect without restart
7. **True Async Execution** - Jobs run independently without blocking the scheduler loop
8. **Parallel Execution** - Jobs execute in parallel using goroutines
9. **Graceful Shutdown** - Waits for all active jobs to complete before exiting
10. **Status Tracking** - Jobs transition through queued → running → completed/failed

## Current State
- **Project folder**: `C:\Users\fe\claudeExperiments\job-scheduler`
- **Status**: Production-ready with web UI and API
- All core features implemented and tested
- **Web Dashboard**: http://localhost:8080 (auto-starts with scheduler)
- **API Endpoints**:
  - GET /api/jobs - List all jobs
  - POST /api/jobs/create - Create new job
  - GET /api/config - Get configuration
  - PUT /api/config - Update configuration
  - GET /api/stats - Get job statistics

## Database Schema

### jobs table
- id (primary key)
- name (job identifier, unique)
- command (shell command to execute)
- status (queued/running/completed/failed)
- last_run (timestamp of last execution)
- created_at, updated_at (timestamps)

### scheduler_config table
- id (primary key)
- key (config name, unique)
- value (config value as text)
- updated_at (timestamp)

**Current config**:
- `max_concurrent_jobs`: Controls how many jobs can run simultaneously (default: 2)

## How It Works
1. Daemon checks database every 10 seconds (configurable via CHECK_INTERVAL env var)
2. Queries `max_concurrent_jobs` from config table
3. Counts currently running jobs (by querying jobs with status='running')
4. Calculates available slots: `max_concurrent_jobs - running_count`
5. If slots available, fetches queued jobs and spawns goroutines (up to available slots)
6. **Scheduler returns immediately** - doesn't wait for jobs to complete
7. Jobs run independently in background goroutines
8. Each job updates its own status (running → completed/failed) in the database
9. Next check cycle sees updated running count and respects the limit
10. On shutdown, scheduler waits for all active goroutines to complete (graceful shutdown)

## Configuration
Edit `.env` file:
```
DB_HOST=localhost
DB_PORT=5432
DB_NAME=scheduler_db
DB_USER=postgres
DB_PASSWORD=postgres
CHECK_INTERVAL=10
```

Change concurrent limit without restart:
```sql
UPDATE scheduler_config SET value='5' WHERE key='max_concurrent_jobs';
```

## Running the Scheduler
```bash
# Build
go build -o job-scheduler.exe

# Run
./job-scheduler.exe
```

## Adding Jobs

### Via Web Dashboard (Easiest)
1. Open http://localhost:8080
2. Fill in the "Add New Job" form
3. Click "Create Job"

### Via API
```bash
curl -X POST http://localhost:8080/api/jobs/create \
  -H "Content-Type: application/json" \
  -d '{"name": "my_job", "command": "echo Hello"}'
```

### Via Database
```sql
INSERT INTO jobs (name, command, status) VALUES
    ('my_job', 'echo "Hello World"', 'queued');
```

## Notes
- Using GenZ communication style as per user preference
- Built for speed and efficiency with Go
- Config changes are picked up automatically on next check cycle
- Jobs execute shell commands - can run any bash/powershell script
