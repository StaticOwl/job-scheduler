# Setup Guide

Detailed instructions for setting up the Job Scheduler Daemon.

## Prerequisites

### 1. Install Go
- Download from: https://go.dev/dl/
- Version required: 1.25.4 or higher
- Verify installation:
  ```bash
  go version
  ```

### 2. Install PostgreSQL
- Download from: https://www.postgresql.org/download/
- Version required: 17.5 or higher
- Verify installation:
  ```bash
  psql --version
  ```

## Database Setup

### Step 1: Create the Database
```bash
# Connect to PostgreSQL
psql -U postgres

# Run the schema file
\i schema.sql

# Run the config migration
\i migration_add_config.sql

# Exit
\q
```

Or run it directly:
```bash
psql -U postgres -f schema.sql
psql -U postgres -f migration_add_config.sql
```

### Step 2: Verify Database Setup
```bash
psql -U postgres -d scheduler_db

# Check tables
\dt

# Verify jobs table
SELECT * FROM jobs;

# Verify config table
SELECT * FROM scheduler_config;
```

You should see:
- `jobs` table with 2 sample jobs
- `scheduler_config` table with `max_concurrent_jobs = 2`

## Application Setup

### Step 1: Configure Environment
```bash
# Copy the example env file
cp .env.example .env

# Edit with your credentials
nano .env  # or use any text editor
```

Update these values in `.env`:
```env
DB_HOST=localhost
DB_PORT=5432
DB_NAME=scheduler_db
DB_USER=postgres
DB_PASSWORD=your_password_here  # CHANGE THIS
CHECK_INTERVAL=10
```

### Step 2: Install Dependencies
```bash
go mod tidy
```

This will download:
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/joho/godotenv` - Environment variable loader

### Step 3: Build the Scheduler
```bash
go build -o job-scheduler.exe ./cmd/scheduler
```

On Linux/Mac:
```bash
go build -o job-scheduler ./cmd/scheduler
```

## Running the Scheduler

### Start the Daemon
```bash
./job-scheduler.exe
```

You should see output like:
```
Starting Job Scheduler Daemon...
Check interval: 10 seconds
Connected to database successfully
Starting web dashboard on http://localhost:8080
Scheduler is running, checking for queued jobs...
Max concurrent: 2, Running: 0, Available slots: 2
Found 2 queued job(s)
...
```

### Access the Web Dashboard
Once the daemon is running:
1. Open your browser
2. Navigate to: **http://localhost:8080**
3. You'll see the Job Scheduler Dashboard with:
   - Real-time job statistics
   - Job table with filtering
   - Form to create new jobs
   - Config management

The dashboard auto-refreshes every 5 seconds!

### Stop the Daemon
Press `Ctrl+C` to gracefully stop the daemon. It will wait for all running jobs to complete before shutting down.

## Verification

### Test Job Execution (Web Dashboard Method)
1. Open the dashboard: http://localhost:8080
2. Fill in the "Add New Job" form:
   - Name: `test_job`
   - Command: `echo "Hello from scheduler"`
3. Click "Create Job"
4. Watch the job appear in the table with status "queued"
5. Within 10 seconds, the status should change to "running" then "completed"
6. The "Completed" stat card should increment

### Test Job Execution (Database Method)
1. Add a test job:
   ```sql
   psql -U postgres -d scheduler_db -c "INSERT INTO jobs (name, command, status) VALUES ('test', 'echo \"Hello from scheduler\"', 'queued');"
   ```

2. Watch the scheduler logs - it should pick up and execute the job within 10 seconds

3. Check job status:
   ```sql
   psql -U postgres -d scheduler_db -c "SELECT name, status, last_run FROM jobs WHERE name='test';"
   ```

### Test Job Execution (API Method)
```bash
# Create a job via API
curl -X POST http://localhost:8080/api/jobs/create \
  -H "Content-Type: application/json" \
  -d '{"name": "api_test", "command": "echo Hello"}'

# Check job status
curl http://localhost:8080/api/jobs | grep api_test
```

### Test Concurrent Limiting
1. Add multiple long-running jobs:
   ```sql
   psql -U postgres -d scheduler_db -c "INSERT INTO jobs (name, command, status) VALUES
     ('sleep1', 'sleep 10 && echo \"Job 1 done\"', 'queued'),
     ('sleep2', 'sleep 10 && echo \"Job 2 done\"', 'queued'),
     ('sleep3', 'sleep 10 && echo \"Job 3 done\"', 'queued');"
   ```

2. Watch the logs - you should see:
   - First 2 jobs execute immediately
   - Third job waits for a slot to open

### Test Dynamic Config (Web Dashboard Method)
1. Open the dashboard: http://localhost:8080
2. In the header, change "Max Concurrent Jobs" from 2 to 5
3. Click "Update"
4. You'll see a success notification
5. Watch the logs or the dashboard - on the next check cycle (within 10 seconds), the new limit takes effect

### Test Dynamic Config (Database Method)
1. While scheduler is running, change the limit:
   ```sql
   psql -U postgres -d scheduler_db -c "UPDATE scheduler_config SET value='5' WHERE key='max_concurrent_jobs';"
   ```

2. Watch the logs - on the next check cycle, you'll see "Max concurrent: 5"

### Test Dynamic Config (API Method)
```bash
curl -X PUT http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{"max_concurrent_jobs": 5}'
```

## Troubleshooting

### "Failed to connect to database"
- Verify PostgreSQL is running: `pg_ctl status`
- Check credentials in `.env` file
- Test connection: `psql -U postgres -d scheduler_db`

### "go: command not found"
- Go is not in your PATH
- Restart your terminal after installing Go
- Verify with: `go version`

### "psql: command not found"
- PostgreSQL is not in your PATH
- On Windows, add `C:\Program Files\PostgreSQL\17\bin` to PATH
- Restart your terminal

### Jobs Not Executing
- Check job status: `SELECT * FROM jobs WHERE status='queued';`
- Verify scheduler is running
- Check logs for errors
- Ensure commands are valid shell commands

### Web Dashboard Not Loading
- Verify scheduler is running
- Check that port 8080 is not in use: `netstat -ano | findstr :8080`
- Try accessing: http://localhost:8080
- Check browser console for errors (F12)
- Verify web/static/ folder exists with HTML/CSS/JS files

### Dashboard Shows "Failed to load jobs"
- Verify scheduler backend is running
- Check API endpoint: `curl http://localhost:8080/api/jobs`
- Check browser console for CORS or network errors
- Verify database connection is working

### API Returns Errors
- Check scheduler logs for detailed error messages
- Verify JSON format in POST requests
- Ensure job names are unique
- Check that max_concurrent_jobs value is a positive integer

## Next Steps

- Read [ARCHITECTURE.md](ARCHITECTURE.md) to understand how it works
- Check [UPCOMING.md](UPCOMING.md) for planned features
- Add your own jobs and start scheduling!
