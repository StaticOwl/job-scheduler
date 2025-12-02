package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type Job struct {
	ID        int
	Name      string
	Command   string
	Status    string
	LastRun   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Database struct {
	conn *sql.DB
}

func New(dsn string) (*Database, error) {
	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Connected to database successfully")
	return &Database{conn: conn}, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
}

func (db *Database) GetQueuedJobs() ([]Job, error) {
	query := `
		SELECT id, name, command, status, last_run, created_at, updated_at
		FROM jobs
		WHERE status = 'queued'
		ORDER BY created_at ASC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query queued jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		err := rows.Scan(&job.ID, &job.Name, &job.Command, &job.Status,
			&job.LastRun, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

func (db *Database) UpdateJobStatus(jobID int, status string) error {
	query := `
		UPDATE jobs
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := db.conn.Exec(query, status, time.Now(), jobID)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}

	return nil
}

func (db *Database) UpdateJobLastRun(jobID int) error {
	query := `
		UPDATE jobs
		SET last_run = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := db.conn.Exec(query, time.Now(), time.Now(), jobID)
	if err != nil {
		return fmt.Errorf("failed to update job last_run: %w", err)
	}

	return nil
}

func (db *Database) GetMaxConcurrentJobs() (int, error) {
	query := `
		SELECT value
		FROM scheduler_config
		WHERE key = 'max_concurrent_jobs'
	`

	var valueStr string
	err := db.conn.QueryRow(query).Scan(&valueStr)
	if err != nil {
		return 0, fmt.Errorf("failed to get max_concurrent_jobs config: %w", err)
	}

	var maxJobs int
	_, err = fmt.Sscanf(valueStr, "%d", &maxJobs)
	if err != nil {
		return 0, fmt.Errorf("failed to parse max_concurrent_jobs value: %w", err)
	}

	return maxJobs, nil
}

func (db *Database) GetRunningJobCount() (int, error) {
	query := `
		SELECT COUNT(*)
		FROM jobs
		WHERE status = 'running'
	`

	var count int
	err := db.conn.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count running jobs: %w", err)
	}

	return count, nil
}

func (db *Database) GetAllJobs() ([]Job, error) {
	query := `
		SELECT id, name, command, status, last_run, created_at, updated_at
		FROM jobs
		ORDER BY created_at DESC
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		err := rows.Scan(&job.ID, &job.Name, &job.Command, &job.Status,
			&job.LastRun, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

func (db *Database) CreateJob(name, command string) error {
	query := `
		INSERT INTO jobs (name, command, status)
		VALUES ($1, $2, 'queued')
	`

	_, err := db.conn.Exec(query, name, command)
	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	log.Printf("Created new job: %s", name)
	return nil
}

func (db *Database) UpdateMaxConcurrentJobs(value int) error {
	query := `
		UPDATE scheduler_config
		SET value = $1, updated_at = $2
		WHERE key = 'max_concurrent_jobs'
	`

	_, err := db.conn.Exec(query, fmt.Sprintf("%d", value), time.Now())
	if err != nil {
		return fmt.Errorf("failed to update max_concurrent_jobs: %w", err)
	}

	log.Printf("Updated max_concurrent_jobs to %d", value)
	return nil
}

func (db *Database) GetJobStats() (map[string]int, error) {
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN status = 'queued' THEN 1 ELSE 0 END) as queued,
			SUM(CASE WHEN status = 'running' THEN 1 ELSE 0 END) as running,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed,
			SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
		FROM jobs
	`

	var total, queued, running, completed, failed int
	err := db.conn.QueryRow(query).Scan(&total, &queued, &running, &completed, &failed)
	if err != nil {
		return nil, fmt.Errorf("failed to get job stats: %w", err)
	}

	return map[string]int{
		"total":     total,
		"queued":    queued,
		"running":   running,
		"completed": completed,
		"failed":    failed,
	}, nil
}
