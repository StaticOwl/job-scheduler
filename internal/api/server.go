package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"job-scheduler/internal/database"
)

type Server struct {
	db   *database.Database
	port int
}

type JobResponse struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	Command   string     `json:"command"`
	Status    string     `json:"status"`
	LastRun   *time.Time `json:"last_run"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateJobRequest struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type ConfigResponse struct {
	MaxConcurrentJobs int `json:"max_concurrent_jobs"`
}

type StatsResponse struct {
	QueuedCount    int `json:"queued_count"`
	RunningCount   int `json:"running_count"`
	CompletedCount int `json:"completed_count"`
	FailedCount    int `json:"failed_count"`
	TotalCount     int `json:"total_count"`
}

func New(db *database.Database, port int) *Server {
	return &Server{
		db:   db,
		port: port,
	}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/jobs", s.handleJobs)
	mux.HandleFunc("/api/jobs/create", s.handleCreateJob)
	mux.HandleFunc("/api/config", s.handleConfig)
	mux.HandleFunc("/api/stats", s.handleStats)

	// Serve static files
	fs := http.FileServer(http.Dir("web/static"))
	mux.Handle("/", fs)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting web dashboard on http://localhost%s", addr)

	return http.ListenAndServe(addr, s.enableCORS(mux))
}

func (s *Server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs, err := s.db.GetAllJobs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]JobResponse, len(jobs))
	for i, job := range jobs {
		response[i] = JobResponse{
			ID:        job.ID,
			Name:      job.Name,
			Command:   job.Command,
			Status:    job.Status,
			LastRun:   job.LastRun,
			CreatedAt: job.CreatedAt,
			UpdatedAt: job.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) handleCreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Command == "" {
		http.Error(w, "Name and command are required", http.StatusBadRequest)
		return
	}

	if err := s.db.CreateJob(req.Name, req.Command); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Job created successfully"})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		maxConcurrent, err := s.db.GetMaxConcurrentJobs()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ConfigResponse{MaxConcurrentJobs: maxConcurrent})

	case http.MethodPut:
		var config ConfigResponse
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		if err := s.db.UpdateMaxConcurrentJobs(config.MaxConcurrentJobs); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Config updated successfully"})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := s.db.GetJobStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(StatsResponse{
		QueuedCount:    stats["queued"],
		RunningCount:   stats["running"],
		CompletedCount: stats["completed"],
		FailedCount:    stats["failed"],
		TotalCount:     stats["total"],
	})
}
