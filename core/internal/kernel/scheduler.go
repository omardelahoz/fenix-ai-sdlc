package kernel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// Scheduler schedules and manages task execution.
type Scheduler struct {
	mu           sync.RWMutex
	queue        []*contracts.ExecutionPackage
	running      map[string]bool
	executor     Executor
	priorityMap map[int]string // Priority -> Provider mapping
	config       *SchedulerConfig
}

// SchedulerConfig contains configuration for the Scheduler.
type SchedulerConfig struct {
	MaxQueueSize    int
	WorkerCount     int
	CheckInterval   time.Duration
}

// DefaultSchedulerConfig returns default scheduler configuration.
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		MaxQueueSize:  1000,
		WorkerCount:   4,
		CheckInterval: 100 * time.Millisecond,
	}
}

// Executor represents something that can execute an ExecutionPackage.
type Executor interface {
	Execute(ctx context.Context, pkg *contracts.ExecutionPackage) (*contracts.ExecutionHandle, error)
}

// NewScheduler creates a new Scheduler.
func NewScheduler(executor Executor) *Scheduler {
	return NewSchedulerWithConfig(executor, DefaultSchedulerConfig())
}

// NewSchedulerWithConfig creates a new Scheduler with custom config.
func NewSchedulerWithConfig(executor Executor, config *SchedulerConfig) *Scheduler {
	return &Scheduler{
		queue:        make([]*contracts.ExecutionPackage, 0, config.MaxQueueSize),
		running:      make(map[string]bool),
		executor:     executor,
		priorityMap: make(map[int]string),
		config:       config,
	}
}

// Schedule adds a task to the scheduler queue.
func (s *Scheduler) Schedule(ctx context.Context, pkg *contracts.ExecutionPackage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate package
	if err := pkg.Validate(); err != nil {
		return fmt.Errorf("package validation failed: %w", err)
	}

	// Check queue size
	if len(s.queue) >= s.config.MaxQueueSize {
		return fmt.Errorf("scheduler queue is full (max %d)", s.config.MaxQueueSize)
	}

	// Add to queue
	s.queue = append(s.queue, pkg)

	return nil
}

// ScheduleDAG schedules multiple packages as a DAG (simplified).
func (s *Scheduler) ScheduleDAG(ctx context.Context, packages []*contracts.ExecutionPackage) error {
	for _, pkg := range packages {
		if err := s.Schedule(ctx, pkg); err != nil {
			return fmt.Errorf("failed to schedule package %s: %w", pkg.Metadata.ID, err)
		}
	}
	return nil
}

// Run starts the scheduler loop.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			s.processQueue(ctx)
		}
	}
}

// processQueue processes queued packages.
func (s *Scheduler) processQueue(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.queue) == 0 {
		return
	}

	// Take the first package
	pkg := s.queue[0]
	s.queue = s.queue[1:]

	// Check if already running
	if s.running[pkg.Metadata.ID] {
		// Skip, already running
		return
	}

	// Mark as running
	s.running[pkg.Metadata.ID] = true

	// Execute in background
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.running, pkg.Metadata.ID)
			s.mu.Unlock()
		}()

		_, err := s.executor.Execute(ctx, pkg)
		if err != nil {
			fmt.Printf("Execution failed for package %s: %v\n", pkg.Metadata.ID, err)
		}
	}()
}

// GetQueueSize returns the current queue size.
func (s *Scheduler) GetQueueSize() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.queue)
}

// GetRunningCount returns the number of currently running tasks.
func (s *Scheduler) GetRunningCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.running)
}

// Cancel cancels a scheduled or running task.
func (s *Scheduler) Cancel(ctx context.Context, packageID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove from queue if present
	for i, pkg := range s.queue {
		if pkg.Metadata.ID == packageID {
			s.queue = append(s.queue[:i], s.queue[i+1:]...)
			return nil
		}
	}

	// If running, the executor should handle cancellation
	if s.running[packageID] {
		return fmt.Errorf("package %s is currently running, cancellation handled by executor", packageID)
	}

	return fmt.Errorf("package %s not found in queue", packageID)
}

// Stats returns scheduler statistics.
func (s *Scheduler) Stats() SchedulerStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return SchedulerStats{
		QueueSize:   len(s.queue),
		RunningCount: len(s.running),
		MaxQueueSize: s.config.MaxQueueSize,
		WorkerCount:  s.config.WorkerCount,
	}
}

// SchedulerStats contains statistics about the scheduler.
type SchedulerStats struct {
	QueueSize   int
	RunningCount int
	MaxQueueSize int
	WorkerCount  int
}
