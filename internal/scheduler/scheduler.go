package scheduler

import (
	"context"
	"log"
	"os/exec"
	"sync"
	"time"

	"job-scheduler/internal/database"
)

type Scheduler struct {
	db              *database.Database
	checkInterval   int
	activeJobs      sync.WaitGroup  // Track all running job goroutines
}

func New(db *database.Database, checkInterval int) *Scheduler {
	return &Scheduler{
		db:            db,
		checkInterval: checkInterval,
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(s.checkInterval) * time.Second)
	defer ticker.Stop()

	log.Println("Scheduler is running, checking for queued jobs...")

	// Run immediately on start
	s.checkAndExecuteJobs()

	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopping, waiting for active jobs to complete...")
			s.Shutdown()
			return
		case <-ticker.C:
			s.checkAndExecuteJobs()
		}
	}
}

// Shutdown waits for all active job goroutines to complete
func (s *Scheduler) Shutdown() {
	s.activeJobs.Wait()
	log.Println("All active jobs completed")
}

func (s *Scheduler) checkAndExecuteJobs() {
	// Get max concurrent jobs from config
	maxConcurrent, err := s.db.GetMaxConcurrentJobs()
	if err != nil {
		log.Printf("Error fetching max_concurrent_jobs config: %v", err)
		return
	}

	// Get current running job count
	runningCount, err := s.db.GetRunningJobCount()
	if err != nil {
		log.Printf("Error counting running jobs: %v", err)
		return
	}

	// Calculate available slots
	availableSlots := maxConcurrent - runningCount
	log.Printf("Max concurrent: %d, Running: %d, Available slots: %d",
		maxConcurrent, runningCount, availableSlots)

	if availableSlots <= 0 {
		log.Println("No available slots, waiting for running jobs to complete")
		return
	}

	// Get queued jobs
	jobs, err := s.db.GetQueuedJobs()
	if err != nil {
		log.Printf("Error fetching queued jobs: %v", err)
		return
	}

	if len(jobs) == 0 {
		log.Println("No queued jobs found")
		return
	}

	log.Printf("Found %d queued job(s)", len(jobs))

	// Only execute up to available slots
	jobsToExecute := jobs
	if len(jobs) > availableSlots {
		jobsToExecute = jobs[:availableSlots]
		log.Printf("Limiting execution to %d job(s) due to concurrent limit", availableSlots)
	}

	// Execute jobs in parallel using goroutines (non-blocking)
	for _, job := range jobsToExecute {
		s.activeJobs.Add(1)
		go func(j database.Job) {
			defer s.activeJobs.Done()
			log.Printf("Executing job: %s (ID: %d)", j.Name, j.ID)
			s.executeJob(j)
		}(job)
	}

	log.Printf("Spawned %d job(s), continuing to next check cycle", len(jobsToExecute))
}

func (s *Scheduler) executeJob(job database.Job) {
	// Update status to running
	if err := s.db.UpdateJobStatus(job.ID, "running"); err != nil {
		log.Printf("Failed to update job %d to running: %v", job.ID, err)
		return
	}

	// Execute the command
	cmd := exec.Command("sh", "-c", job.Command)
	output, err := cmd.CombinedOutput()

	// Update last run time
	if err := s.db.UpdateJobLastRun(job.ID); err != nil {
		log.Printf("Failed to update last_run for job %d: %v", job.ID, err)
	}

	if err != nil {
		log.Printf("Job %s failed: %v\nOutput: %s", job.Name, err, string(output))
		if err := s.db.UpdateJobStatus(job.ID, "failed"); err != nil {
			log.Printf("Failed to update job %d to failed: %v", job.ID, err)
		}
		return
	}

	log.Printf("Job %s completed successfully\nOutput: %s", job.Name, string(output))

	// Update status to completed
	if err := s.db.UpdateJobStatus(job.ID, "completed"); err != nil {
		log.Printf("Failed to update job %d to completed: %v", job.ID, err)
	}
}
