package kernel

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

func TestFenixKernel_ConcurrentExecution(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Initialize
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer k.Shutdown(ctx)

	// Execute multiple packages concurrently
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			pkg := &contracts.ExecutionPackage{
				Metadata: contracts.PackageMetadata{
					ID:            "pkg-concurrent-" + string(rune('0'+id)),
					Version:       "v1",
					SchemaVersion: "1.0",
					KernelVersion: "0.1.0",
					TraceID:       "trace-" + string(rune('0'+id)),
					CorrelationID: "corr-" + string(rune('0'+id)),
				},
				Task: contracts.TaskDescriptor{
					ID:   "task-concurrent-" + string(rune('0'+id)),
					Type: "GenerateCode",
				},
				Executor: contracts.ExecutorDescriptor{
					ID:           "executor-concurrent-" + string(rune('0'+id)),
					Type:         "Agent",
					Version:      "1.0",
					Capabilities: []string{"code_generation"},
				},
				Snapshot: contracts.ExecutionSnapshot{
					AstNodes:         []string{"node-1", "node-2"},
					MemoryReferences: []string{"ref-1", "ref-2"},
				},
				Policies: contracts.ExecutionPolicies{
					MaxTokens:           100000,
					MaxFiles:            100,
					MaxAstDepth:         50,
					CompressionStrategy: "gzip",
					RetryPolicy: contracts.RetryConfig{
						MaxRetries:      3,
						BackoffStrategy: "exponential",
					},
				},
				Resources: contracts.ResourceAllocation{
					Provider:        "claude-3.5-sonnet",
					EstimatedTokens: 50000,
					ReservedTokens:  60000,
					EstimatedCost:   0.15,
					Priority:        1,
					Deadline:        time.Now().Add(30 * time.Minute),
				},
				Contracts: contracts.Contracts{
					ExpectedFormat:  "json",
					JsonSchema:      "{}",
					ValidationRules: []string{"required_fields", "type_check"},
				},
			}

			_, err := k.Execute(ctx, pkg)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent execution failed: %v", err)
	}

	// Verify all handles were created
	handles := k.ListHandles()
	if len(handles) != 5 {
		t.Errorf("Expected 5 handles, got %d", len(handles))
	}
}

func TestFenixKernel_ConcurrentListHandles(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Initialize
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer k.Shutdown(ctx)

	// Execute some packages first
	for i := 0; i < 3; i++ {
		pkg := &contracts.ExecutionPackage{
			Metadata: contracts.PackageMetadata{
				ID: "pkg-list-concurrent-" + string(rune('0'+i)),
			},
			Task: contracts.TaskDescriptor{
				ID: "task-list-concurrent-" + string(rune('0'+i)),
			},
			Executor: contracts.ExecutorDescriptor{
				ID: "executor-list-concurrent-" + string(rune('0'+i)),
			},
			Resources: contracts.ResourceAllocation{
				Provider: "claude",
				Deadline: time.Now().Add(time.Hour),
			},
		}

		_, err := k.Execute(ctx, pkg)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	}

	// List handles concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			handles := k.ListHandles()
			if len(handles) != 3 {
				t.Errorf("Expected 3 handles, got %d", len(handles))
			}
		}()
	}

	wg.Wait()
}

func TestFenixKernel_ConcurrentStateTransitions(t *testing.T) {
	k := NewKernel()

	transitionCount := 0
	var mu sync.Mutex

	k.OnTransition(func(from, to contracts.KernelState) {
		mu.Lock()
		transitionCount++
		mu.Unlock()
	})

	ctx := context.Background()

	// Multiple goroutines trying to initialize
	var wg sync.WaitGroup
	initErrors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := k.Initialize(ctx)
			if err != nil {
				initErrors <- err
			}
		}()
	}

	wg.Wait()
	close(initErrors)

	// Only first initialize should succeed, others should fail
	errorCount := 0
	for err := range initErrors {
		if err != nil {
			errorCount++
		}
	}

	if errorCount != 4 {
		t.Errorf("Expected 4 initialization errors, got %d", errorCount)
	}

	// Should have exactly 2 transitions (Uninitialized → Initializing → Running)
	mu.Lock()
	count := transitionCount
	mu.Unlock()

	if count != 2 {
		t.Errorf("Expected 2 state transitions, got %d", count)
	}
}

func TestExecutionManager_ConcurrentCreateHandle(t *testing.T) {
	mgr := NewExecutionManager()

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			pkg := &contracts.ExecutionPackage{
				Metadata: contracts.PackageMetadata{
					ID: "pkg-mgr-concurrent-" + string(rune('0'+id)),
				},
				Task: contracts.TaskDescriptor{
					ID: "task-mgr-concurrent-" + string(rune('0'+id)),
				},
				Executor: contracts.ExecutorDescriptor{
					ID: "executor-mgr-concurrent-" + string(rune('0'+id)),
				},
				Resources: contracts.ResourceAllocation{
					Provider: "claude",
					Deadline: time.Now().Add(time.Hour),
				},
			}

			_, err := mgr.CreateHandle(pkg)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent CreateHandle failed: %v", err)
	}

	// Verify all handles were created
	handles := mgr.ListHandles()
	if len(handles) != 10 {
		t.Errorf("Expected 10 handles, got %d", len(handles))
	}
}

func TestMemoryManager_ConcurrentCache(t *testing.T) {
	mgr := NewMemoryManager()

	ctx := context.Background()

	snapshot := &contracts.ExecutionSnapshot{
		AstNodes:         []string{"node-1", "node-2"},
		MemoryReferences: []string{"ref-1", "ref-2"},
	}

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Concurrent cache operations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			snapshotID := "snapshot-concurrent-" + string(rune('0'+id))
			err := mgr.Cache(ctx, snapshotID, snapshot)
			if err != nil {
				errors <- err
			}

			// Also try to get it
			_, err = mgr.Get(ctx, snapshotID)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent cache operation failed: %v", err)
	}

	// Verify all snapshots were cached
	stats := mgr.Stats()
	if stats.SnapshotCount != 10 {
		t.Errorf("Expected 10 snapshots cached, got %d", stats.SnapshotCount)
	}
}

func TestMemoryManager_ConcurrentResolve(t *testing.T) {
	mgr := NewMemoryManager()

	ctx := context.Background()

	snapshot := &contracts.ExecutionSnapshot{
		AstNodes:         []string{"node-1", "node-2", "node-3"},
		MemoryReferences: []string{"ref-1", "ref-2", "ref-3"},
	}

	// Cache the snapshot first
	if err := mgr.Cache(ctx, "snapshot-resolve", snapshot); err != nil {
		t.Fatalf("Cache failed: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := mgr.Resolve(ctx, snapshot)
			if err != nil {
				t.Errorf("Resolve failed: %v", err)
			}
		}()
	}

	wg.Wait()

	// Verify cache hits
	stats := mgr.Stats()
	if stats.Hits < 20 {
		t.Errorf("Expected at least 20 cache hits, got %d", stats.Hits)
	}
}

func TestScheduler_ConcurrentSchedule(t *testing.T) {
	// Create a mock executor
	executor := &mockExecutor{
		executeFunc: func(ctx context.Context, pkg *contracts.ExecutionPackage) (*contracts.ExecutionHandle, error) {
			time.Sleep(10 * time.Millisecond) // Simulate work
			return contracts.NewExecutionHandle(pkg), nil
		},
	}

	scheduler := NewScheduler(executor)

	ctx := context.Background()

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Schedule multiple packages concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			pkg := &contracts.ExecutionPackage{
				Metadata: contracts.PackageMetadata{
					ID: "pkg-scheduler-" + string(rune('0'+id)),
				},
				Task: contracts.TaskDescriptor{
					ID: "task-scheduler-" + string(rune('0'+id)),
				},
				Executor: contracts.ExecutorDescriptor{
					ID: "executor-scheduler-" + string(rune('0'+id)),
				},
				Resources: contracts.ResourceAllocation{
					Provider: "claude",
					Deadline: time.Now().Add(time.Hour),
				},
			}

			err := scheduler.Schedule(ctx, pkg)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent schedule failed: %v", err)
	}

	// Verify all packages were scheduled
	stats := scheduler.Stats()
	if stats.QueueSize != 10 {
		t.Errorf("Expected 10 packages in queue, got %d", stats.QueueSize)
	}
}

// mockExecutor is a mock implementation of Executor for testing.
type mockExecutor struct {
	executeFunc func(ctx context.Context, pkg *contracts.ExecutionPackage) (*contracts.ExecutionHandle, error)
}

func (m *mockExecutor) Execute(ctx context.Context, pkg *contracts.ExecutionPackage) (*contracts.ExecutionHandle, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, pkg)
	}
	return contracts.NewExecutionHandle(pkg), nil
}
