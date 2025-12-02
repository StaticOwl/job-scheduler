.PHONY: build run test clean help

# Build the scheduler
build:
	go build -o job-scheduler.exe ./cmd/scheduler

# Run the scheduler
run: build
	./job-scheduler.exe

# Run all tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run only unit tests (excluding integration tests)
test-unit:
	go test ./internal/...

# Clean build artifacts
clean:
	rm -f job-scheduler.exe
	rm -f job-scheduler

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build the scheduler binary"
	@echo "  run            - Build and run the scheduler"
	@echo "  test           - Run all tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-unit      - Run only unit tests"
	@echo "  clean          - Remove build artifacts"
	@echo "  help           - Show this help message"
