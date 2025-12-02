# Job Scheduler Daemon

A lightweight, high-performance job scheduler daemon built in Go that monitors and executes jobs from a PostgreSQL database with configurable concurrent execution limits.

Built with Claude Code as a part of Experiment. The goals are:
>* Capabilities of AI Tools to generate project.
>* Extent of features can be written with AI Tools.
>* Extent of Customisation can be done to AI Tools.
>* Accuracy of AI Tools.

Ultimate Purpose: At least automate 70-80% of code generation with AI Tools, if working.

## Features

- **Web Dashboard** - Real-time job visualization and management UI
- **REST API** - Full API for programmatic job control
- **Continuous Job Monitoring** - Daemon runs 24/7 checking for queued jobs
- **Concurrent Execution Control** - Limit how many jobs run simultaneously
- **Dynamic Configuration** - Change settings without restarting the daemon
- **Parallel Processing** - Execute multiple jobs in parallel with goroutines
- **Status Tracking** - Full job lifecycle tracking (queued → running → completed/failed)
- **Database-Driven** - All config and jobs stored in PostgreSQL

## Quick Start

### Prerequisites
- Go 1.25.4 or higher
- PostgreSQL 17.5 or higher

### Setup
```bash
# 1. Clone or navigate to the project
cd job-scheduler

# 2. Set up the database
psql -U postgres -f schema.sql
psql -U postgres -f migration_add_config.sql

# 3. Configure environment
cp .env.example .env
# Edit .env with your database credentials

# 4. Install dependencies
go mod tidy

# 5. Build the scheduler
go build -o job-scheduler.exe ./cmd/scheduler

# Or use make
make build

# 6. Run it
./job-scheduler.exe

# Or use make
make run

# 7. Access the web dashboard
# Open your browser and go to: http://localhost:8080
```

## How to Use

### Web Dashboard (Recommended)

1. **Start the scheduler**: `./job-scheduler.exe`
2. **Open your browser**: Navigate to `http://localhost:8080`
3. **Monitor jobs**: See real-time stats and job status
4. **Add jobs**: Use the form on the dashboard
5. **Update config**: Change max concurrent jobs from the UI

The dashboard auto-refreshes every 5 seconds!

### Adding Jobs via Database
```sql
INSERT INTO jobs (name, command, status) VALUES
    ('backup_job', 'pg_dump mydb > backup.sql', 'queued');
```

### Adding Jobs via API
```bash
curl -X POST http://localhost:8080/api/jobs/create \
  -H "Content-Type: application/json" \
  -d '{"name": "my_job", "command": "echo Hello"}'
```

### Changing Concurrent Limit
**Via Dashboard**: Update the value in the header and click "Update"

**Via Database**:
```sql
UPDATE scheduler_config SET value='5' WHERE key='max_concurrent_jobs';
```

**Via API**:
```bash
curl -X PUT http://localhost:8080/api/config \
  -H "Content-Type: application/json" \
  -d '{"max_concurrent_jobs": 5}'
```
Changes take effect on the next check cycle (no restart needed!)

### Monitoring Jobs
```sql
-- See all jobs
SELECT * FROM jobs;

-- See running jobs
SELECT * FROM jobs WHERE status='running';

-- See completed jobs
SELECT * FROM jobs WHERE status='completed' ORDER BY last_run DESC;
```

## Configuration

Environment variables (`.env` file):
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_NAME` - Database name (default: scheduler_db)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password
- `CHECK_INTERVAL` - Seconds between checks (default: 10)

Database config (`scheduler_config` table):
- `max_concurrent_jobs` - Maximum jobs running simultaneously (default: 2)

## Project Structure

```
job-scheduler/
├── README.md                    # This file
├── CLAUDE.md                    # Project context for Claude Code
├── Makefile                     # Build automation
├── docs/                        # Detailed documentation
│   ├── SETUP.md                # Detailed setup instructions
│   ├── ARCHITECTURE.md         # System architecture
│   └── UPCOMING.md             # Roadmap and upcoming features
├── cmd/                         # Application entry points
│   └── scheduler/
│       └── main.go             # Main daemon entry point
├── internal/                    # Internal packages
│   ├── config/
│   │   ├── config.go           # Configuration loader
│   │   └── config_test.go      # Config tests
│   ├── database/
│   │   └── database.go         # Database operations
│   └── scheduler/
│       └── scheduler.go        # Scheduler logic
├── tests/                       # Integration tests
│   └── integration_test.go     # Integration test suite
├── schema.sql                   # Initial database schema
├── migration_add_config.sql    # Config table migration
├── .env                         # Environment config (not in git)
├── .env.example                # Environment config template
├── .gitignore                  # Git ignore rules
└── go.mod                       # Go dependencies
```

## Documentation

- **[Setup Guide](docs/SETUP.md)** - Detailed installation and configuration
- **[Architecture](docs/ARCHITECTURE.md)** - How the system works
- **[Upcoming Features](docs/UPCOMING.md)** - Roadmap and future plans

## Development

### Running Tests
```bash
# Run all tests
go test ./...

# Or use make
make test

# Run tests with coverage
make test-coverage

# Run only unit tests
make test-unit
```

### Building
```bash
# Build the binary
go build -o job-scheduler.exe ./cmd/scheduler

# Or use make
make build

# Build and run
make run
```

### Available Make Commands
```bash
make help          # Show all available commands
make build         # Build the scheduler
make run           # Build and run
make test          # Run all tests
make test-verbose  # Run tests with verbose output
make test-coverage # Run tests with coverage
make test-unit     # Run only unit tests
make clean         # Remove build artifacts
```

## Tech Stack

- **Language**: Go 1.25.4
- **Database**: PostgreSQL 17.5
- **Platform**: Windows (cross-platform compatible)

## License

MIT

## Contributing

Feel free to open issues or submit PRs!
