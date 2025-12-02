# Architecture

Understanding how the Job Scheduler Daemon works internally.

## System Overview

```
        ┌──────────────────────────────────────────┐
        │          Browser (User)                  │
        │    http://localhost:8080                 │
        └────────────────┬─────────────────────────┘
                         │ HTTP
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                   Job Scheduler Daemon                       │
│                                                              │
│  ┌──────────────────────┐    ┌───────────────────────────┐ │
│  │   REST API Server    │    │   Scheduler Loop          │ │
│  │   (Port 8080)        │    │   (every 10 seconds)      │ │
│  │                      │    │                           │ │
│  │  GET  /api/jobs      │    │  1. Load max_concurrent   │ │
│  │  POST /api/jobs/     │    │  2. Count running jobs    │ │
│  │       create         │    │  3. Calculate slots       │ │
│  │  GET  /api/config    │    │  4. Fetch queued jobs     │ │
│  │  PUT  /api/config    │    │  5. Spawn goroutines      │ │
│  │  GET  /api/stats     │    │  6. Return immediately    │ │
│  │                      │    │     (non-blocking)        │ │
│  │  Serves web/static/  │    │                           │ │
│  └──────────┬───────────┘    └───────────┬───────────────┘ │
│             │                            │                  │
└─────────────┴────────────────────────────┴──────────────────┘
              │                            │
              │        Postgres Connection │
              └────────────────────────────┘
                         │
                         ▼
                ┌──────────────────────┐
                │   PostgreSQL Database │
                │                       │
                │  ┌─────────────────┐ │
                │  │  jobs table      │ │
                │  │  - queued        │ │
                │  │  - running       │ │
                │  │  - completed     │ │
                │  │  - failed        │ │
                │  └─────────────────┘ │
                │                       │
                │  ┌─────────────────┐ │
                │  │ scheduler_config │ │
                │  │ - max_concurrent │ │
                │  └─────────────────┘ │
                └──────────────────────┘
```

## Core Components

### 1. Main Entry Point (`cmd/scheduler/main.go`)
The heart of the application that coordinates all components.

**Responsibilities:**
- Initialize database connection
- Start REST API server in background
- Set up graceful shutdown handling
- Run the scheduler loop
- Handle OS signals (Ctrl+C)

**Startup Sequence:**
1. Load `.env` file
2. Connect to database
3. Start API server (goroutine on port 8080)
4. Create scheduler instance
5. Run scheduler loop (blocks until shutdown)

### 2. REST API Server (`internal/api/server.go`)
HTTP server providing web dashboard and REST API.

**Responsibilities:**
- Serve web dashboard (static files)
- Provide REST API endpoints
- Handle CORS for frontend
- Process HTTP requests

**API Endpoints:**
- `GET /` - Serve web dashboard
- `GET /api/jobs` - List all jobs
- `POST /api/jobs/create` - Create new job
- `GET /api/config` - Get configuration
- `PUT /api/config` - Update configuration
- `GET /api/stats` - Get job statistics

**Port:** 8080 (hardcoded, runs in separate goroutine)

### 3. Scheduler (`internal/scheduler/scheduler.go`)
Core scheduling logic that manages job execution.

**Responsibilities:**
- Run check cycle every N seconds
- Query jobs and config from database
- Respect concurrent job limits
- Spawn job goroutines (non-blocking)
- Track active goroutines
- Graceful shutdown

**Key Functions:**
- `Run()` - Main scheduler loop with ticker
- `checkAndExecuteJobs()` - Check slots and spawn jobs
- `executeJob()` - Execute single job in goroutine
- `Shutdown()` - Wait for all active jobs

### 4. Configuration (`internal/config/config.go`)
Loads environment variables and provides database connection string.

**Responsibilities:**
- Read `.env` file
- Provide default values
- Generate PostgreSQL DSN (Data Source Name)

**Configuration Options:**
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
- `CHECK_INTERVAL` - seconds between checks

### 5. Database Layer (`internal/database/database.go`)
Handles all database operations.

**Key Functions:**
- `New()` - Creates and verifies database connection
- `GetQueuedJobs()` - Fetches jobs with status='queued'
- `GetAllJobs()` - Fetches all jobs for API
- `GetMaxConcurrentJobs()` - Reads config from scheduler_config table
- `GetRunningJobCount()` - Counts jobs with status='running'
- `CreateJob()` - Creates new job via API
- `UpdateMaxConcurrentJobs()` - Updates config via API
- `GetJobStats()` - Returns job counts by status
- `UpdateJobStatus()` - Changes job status
- `UpdateJobLastRun()` - Updates last_run timestamp

### 6. Web Dashboard (`web/static/`)
Frontend interface for job visualization and management.

**Files:**
- `index.html` - Dashboard UI structure
- `style.css` - Styling with gradient background
- `app.js` - Frontend logic and API calls

**Features:**
- **Stats Cards** - Real-time counts (queued/running/completed/failed)
- **Job Table** - Filterable list of all jobs with status badges
- **Create Form** - Add new jobs with name and command
- **Config Editor** - Update max concurrent jobs
- **Auto-Refresh** - Polls API every 5 seconds for updates

**UI Design:**
- Responsive layout (works on mobile)
- Color-coded status badges
- Gradient purple/violet theme
- Toast notifications for actions
- No build process required (vanilla HTML/CSS/JS)

**API Integration:**
```javascript
// Fetch jobs
fetch('http://localhost:8080/api/jobs')

// Create job
fetch('http://localhost:8080/api/jobs/create', {
  method: 'POST',
  body: JSON.stringify({name, command})
})

// Update config
fetch('http://localhost:8080/api/config', {
  method: 'PUT',
  body: JSON.stringify({max_concurrent_jobs})
})
```

## Job Lifecycle

```
┌─────────┐
│ queued  │ ← Job inserted by user
└────┬────┘
     │
     ▼
┌─────────┐
│ running │ ← Daemon picks it up and starts execution
└────┬────┘
     │
     ├─────────────┬─────────────┐
     │             │             │
     ▼             ▼             ▼
┌───────────┐ ┌─────────┐ ┌─────────┐
│ completed │ │ failed  │ │ running │ (if long-running)
└───────────┘ └─────────┘ └─────────┘
```

## Concurrency Management

### How It Works
1. **Read Config**: Query `max_concurrent_jobs` from database
2. **Count Running**: Count jobs where status='running'
3. **Calculate Slots**: `available_slots = max_concurrent - running_count`
4. **Limit Execution**: Only execute up to `available_slots` jobs
5. **Parallel Execution**: Use goroutines with WaitGroup for parallel processing

### Example Flow

**Scenario**: max_concurrent_jobs = 2, 5 jobs queued

```
Check 1 (t=0s):
  Running: 0, Available: 2
  Execute: job1, job2 (in parallel)
  Remaining: job3, job4, job5

Check 2 (t=10s):
  Running: 2, Available: 0
  Execute: nothing (waiting for slots)
  Remaining: job3, job4, job5

Check 3 (t=20s):
  Running: 0, Available: 2 (job1, job2 completed)
  Execute: job3, job4 (in parallel)
  Remaining: job5

Check 4 (t=30s):
  Running: 2, Available: 0
  Execute: nothing
  Remaining: job5

Check 5 (t=40s):
  Running: 0, Available: 2 (job3, job4 completed)
  Execute: job5
  Remaining: none
```

## Parallel Execution

Jobs are executed in parallel using Go's goroutines:

```go
var wg sync.WaitGroup
for _, job := range jobsToExecute {
    wg.Add(1)
    go func(j Job) {
        defer wg.Done()
        executeJob(db, j)
    }(job)
}
wg.Wait()  // Wait for all jobs to complete
```

**Benefits:**
- Multiple jobs run simultaneously
- Efficient resource utilization
- Non-blocking execution

## Dynamic Configuration Reload

The scheduler checks the config table on **every loop iteration**:

```go
maxConcurrent, err := db.GetMaxConcurrentJobs()  // Fresh read from DB
```

**No restart required!** Changes take effect on the next check cycle (within 10 seconds by default).

## Database Schema

### jobs Table
```sql
CREATE TABLE jobs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    command TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'queued',
    last_run TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_jobs_status ON jobs(status);
```

### scheduler_config Table
```sql
CREATE TABLE scheduler_config (
    id SERIAL PRIMARY KEY,
    key VARCHAR(255) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_config_key ON scheduler_config(key);
```

## Graceful Shutdown

The daemon handles shutdown signals properly:

```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigChan
    log.Println("Received shutdown signal, stopping scheduler...")
    cancel()  // Cancels the context
}()
```

When you press Ctrl+C:
1. Signal is caught
2. Context is cancelled
3. Main loop exits after current iteration
4. Database connection closes
5. Program exits cleanly

## Performance Considerations

### Efficient Queries
- Index on `status` column for fast filtering
- Index on `key` column in config table
- Simple COUNT(*) query for running jobs

### Connection Pooling
Uses `database/sql` which includes built-in connection pooling.

### Goroutine Usage
- Creates goroutines per batch, not per job over all time
- WaitGroup ensures proper cleanup
- No goroutine leaks

## Error Handling

The scheduler is resilient:
- Database errors are logged but don't crash the daemon
- Job execution errors are caught and marked as 'failed'
- Config read errors use previous value or skip iteration
- Graceful degradation on issues

## Logging

All operations are logged:
- Startup and shutdown
- Database connection status
- Jobs found and executed
- Success/failure of each job
- Configuration changes detected

Example logs:
```
2025/12/01 21:44:58 Starting Job Scheduler Daemon...
2025/12/01 21:44:58 Check interval: 10 seconds
2025/12/01 21:44:58 Connected to database successfully
2025/12/01 21:44:58 Starting web dashboard on http://localhost:8080
2025/12/01 21:44:58 Scheduler is running, checking for queued jobs...
2025/12/01 21:44:58 Max concurrent: 2, Running: 0, Available slots: 2
2025/12/01 21:44:58 Found 5 queued job(s)
2025/12/01 21:44:58 Limiting execution to 2 job(s) due to concurrent limit
2025/12/01 21:44:58 Spawned 2 job(s), continuing to next check cycle
2025/12/01 21:44:58 Executing job: test_job_1 (ID: 1)
2025/12/01 21:44:58 Job test_job_1 completed successfully
2025/12/01 21:45:02 Created new job: my_new_job
2025/12/01 21:45:15 Updated max_concurrent_jobs to 5
```

## Extensibility

The architecture is designed for easy extension:
- Add more config options in `scheduler_config` table
- Extend `Job` struct with additional fields
- Add more database operations in `database.go`
- Implement job prioritization
- Add retry logic
- Integrate with external systems

See [UPCOMING.md](UPCOMING.md) for planned features.
