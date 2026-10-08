package kernel

import (
	"context"
	"testing"
	"time"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

func TestFenixKernel_Initialize(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Test initialize from uninitialized state
	err := k.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	if k.State() != contracts.StateRunning {
		t.Errorf("Expected state Running, got %s", k.State())
	}

	// Test double initialize should fail
	err = k.Initialize(ctx)
	if err == nil {
		t.Error("Expected error when initializing already initialized kernel")
	}
}

func TestFenixKernel_Shutdown(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Initialize first
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Shutdown
	if err := k.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if k.State() != contracts.StateShutdown {
		t.Errorf("Expected state Shutdown, got %s", k.State())
	}

	// Double shutdown should be safe
	if err := k.Shutdown(ctx); err != nil {
		t.Errorf("Double shutdown should be safe, got error: %v", err)
	}
}

func TestFenixKernel_Execute(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Initialize
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer k.Shutdown(ctx)

	// Create a valid execution package
	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			ID:            "pkg-test-001",
			Version:       "v1",
			SchemaVersion: "1.0",
			KernelVersion: "0.1.0",
			TraceID:       "trace-123",
			CorrelationID: "corr-456",
		},
		Task: contracts.TaskDescriptor{
			ID:   "task-test-001",
			Type: "GenerateCode",
		},
		Executor: contracts.ExecutorDescriptor{
			ID:           "executor-test-001",
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

	// Execute
	handle, err := k.Execute(ctx, pkg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify handle
	if handle == nil {
		t.Fatal("Expected non-nil handle")
	}

	if handle.PackageID != "pkg-test-001" {
		t.Errorf("Expected PackageID pkg-test-001, got %s", handle.PackageID)
	}

	if handle.CurrentVersion != "v1" {
		t.Errorf("Expected CurrentVersion v1, got %s", handle.CurrentVersion)
	}

	if handle.Status != "Completed" {
		t.Errorf("Expected Status Completed, got %s", handle.Status)
	}
}

func TestFenixKernel_Execute_InvalidPackage(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Initialize
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer k.Shutdown(ctx)

	// Create an invalid package (missing required fields)
	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			// Missing ID
		},
		Task: contracts.TaskDescriptor{
			// Missing ID
		},
	}

	// Execute should fail validation
	_, err := k.Execute(ctx, pkg)
	if err == nil {
		t.Error("Expected error for invalid package")
	}
}

func TestFenixKernel_Execute_NotInitialized(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			ID: "pkg-001",
		},
		Task: contracts.TaskDescriptor{
			ID: "task-001",
		},
		Executor: contracts.ExecutorDescriptor{
			ID: "executor-001",
		},
		Resources: contracts.ResourceAllocation{
			Provider: "claude",
			Deadline: time.Now().Add(time.Hour),
		},
	}

	// Execute without initialize should fail
	_, err := k.Execute(ctx, pkg)
	if err == nil {
		t.Error("Expected error when kernel not initialized")
	}
}

func TestFenixKernel_Cancel(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Initialize
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer k.Shutdown(ctx)

	// Create and execute a package
	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			ID: "pkg-cancel-001",
		},
		Task: contracts.TaskDescriptor{
			ID: "task-cancel-001",
		},
		Executor: contracts.ExecutorDescriptor{
			ID: "executor-cancel-001",
		},
		Resources: contracts.ResourceAllocation{
			Provider: "claude",
			Deadline: time.Now().Add(time.Hour),
		},
	}

	handle, err := k.Execute(ctx, pkg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Cancel the execution
	err = k.Cancel(ctx, "pkg-cancel-001")
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	// Verify cancellation (should be handled by the handle)
	select {
	case <-handle.CancellationToken:
		// Channel should be closed
	default:
		t.Error("Expected cancellation channel to be closed")
	}
}

func TestFenixKernel_GetHandle(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Initialize
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer k.Shutdown(ctx)

	// Execute a package
	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			ID: "pkg-get-001",
		},
		Task: contracts.TaskDescriptor{
			ID: "task-get-001",
		},
		Executor: contracts.ExecutorDescriptor{
			ID: "executor-get-001",
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

	// Get the handle
	handle, err := k.GetHandle("pkg-get-001")
	if err != nil {
		t.Fatalf("GetHandle failed: %v", err)
	}

	if handle.PackageID != "pkg-get-001" {
		t.Errorf("Expected PackageID pkg-get-001, got %s", handle.PackageID)
	}
}

func TestFenixKernel_GetHandle_NotFound(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Initialize
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer k.Shutdown(ctx)

	// Try to get non-existent handle
	_, err := k.GetHandle("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent handle")
	}
}

func TestFenixKernel_ListHandles(t *testing.T) {
	k := NewKernel()

	ctx := context.Background()

	// Initialize
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}
	defer k.Shutdown(ctx)

	// Execute multiple packages
	for i := 0; i < 3; i++ {
		pkg := &contracts.ExecutionPackage{
			Metadata: contracts.PackageMetadata{
				ID: "pkg-list-" + string(rune('0'+i)),
			},
			Task: contracts.TaskDescriptor{
				ID: "task-list-" + string(rune('0'+i)),
			},
			Executor: contracts.ExecutorDescriptor{
				ID: "executor-list-" + string(rune('0'+i)),
			},
			Resources: contracts.ResourceAllocation{
				Provider: "claude",
				Deadline: time.Now().Add(time.Hour),
			},
		}

		_, err := k.Execute(ctx, pkg)
		if err != nil {
			t.Fatalf("Execute failed for package %d: %v", i, err)
		}
	}

	// List handles
	handles := k.ListHandles()
	if len(handles) != 3 {
		t.Errorf("Expected 3 handles, got %d", len(handles))
	}
}

func TestFenixKernel_StateTransitions(t *testing.T) {
	k := NewKernel()

	transitionCount := 0
	k.OnTransition(func(from, to contracts.KernelState) {
		transitionCount++
	})

	ctx := context.Background()

	// Initialize should trigger transitions
	if err := k.Initialize(ctx); err != nil {
		t.Fatalf("Initialize failed: %v", err)
	}

	// Should have transitioned from Uninitialized → Initializing → Running
	if transitionCount < 2 {
		t.Errorf("Expected at least 2 transitions, got %d", transitionCount)
	}

	// Shutdown should trigger transitions
	if err := k.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if k.State() != contracts.StateShutdown {
		t.Errorf("Expected state Shutdown, got %s", k.State())
	}
}
