package kernel

import (
	"context"
	"testing"
	"time"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

func TestExecutionManagerImpl_CreateHandle(t *testing.T) {
	mgr := NewExecutionManager()

	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			ID: "pkg-mgr-001",
		},
		Task: contracts.TaskDescriptor{
			ID: "task-mgr-001",
		},
		Executor: contracts.ExecutorDescriptor{
			ID: "executor-mgr-001",
		},
		Resources: contracts.ResourceAllocation{
			Provider: "claude",
			Deadline: time.Now().Add(time.Hour),
		},
	}

	handle, err := mgr.CreateHandle(pkg)
	if err != nil {
		t.Fatalf("CreateHandle failed: %v", err)
	}

	if handle == nil {
		t.Fatal("Expected non-nil handle")
	}

	if handle.PackageID != "pkg-mgr-001" {
		t.Errorf("Expected PackageID pkg-mgr-001, got %s", handle.PackageID)
	}

	if handle.Status != "Running" {
		t.Errorf("Expected Status Running, got %s", handle.Status)
	}
}

func TestExecutionManagerImpl_CreateHandle_Duplicate(t *testing.T) {
	mgr := NewExecutionManager()

	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			ID: "pkg-dup-001",
		},
		Task: contracts.TaskDescriptor{
			ID: "task-dup-001",
		},
		Executor: contracts.ExecutorDescriptor{
			ID: "executor-dup-001",
		},
		Resources: contracts.ResourceAllocation{
			Provider: "claude",
			Deadline: time.Now().Add(time.Hour),
		},
	}

	// Create first handle
	_, err := mgr.CreateHandle(pkg)
	if err != nil {
		t.Fatalf("First CreateHandle failed: %v", err)
	}

	// Try to create duplicate
	_, err = mgr.CreateHandle(pkg)
	if err == nil {
		t.Error("Expected error for duplicate handle")
	}
}

func TestExecutionManagerImpl_GetHandle(t *testing.T) {
	mgr := NewExecutionManager()

	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			ID: "pkg-get-002",
		},
		Task: contracts.TaskDescriptor{
			ID: "task-get-002",
		},
		Executor: contracts.ExecutorDescriptor{
			ID: "executor-get-002",
		},
		Resources: contracts.ResourceAllocation{
			Provider: "claude",
			Deadline: time.Now().Add(time.Hour),
		},
	}

	_, err := mgr.CreateHandle(pkg)
	if err != nil {
		t.Fatalf("CreateHandle failed: %v", err)
	}

	handle, err := mgr.GetHandle("pkg-get-002")
	if err != nil {
		t.Fatalf("GetHandle failed: %v", err)
	}

	if handle.PackageID != "pkg-get-002" {
		t.Errorf("Expected PackageID pkg-get-002, got %s", handle.PackageID)
	}
}

func TestExecutionManagerImpl_GetHandle_NotFound(t *testing.T) {
	mgr := NewExecutionManager()

	_, err := mgr.GetHandle("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent handle")
	}
}

func TestExecutionManagerImpl_UpdateHandle(t *testing.T) {
	mgr := NewExecutionManager()

	pkg := &contracts.ExecutionPackage{
		Metadata: contracts.PackageMetadata{
			ID: "pkg-update-001",
		},
		Task: contracts.TaskDescriptor{
			ID: "task-update-001",
		},
		Executor: contracts.ExecutorDescriptor{
			ID: "executor-update-001",
		},
		Resources: contracts.ResourceAllocation{
			Provider: "claude",
			Deadline: time.Now().Add(time.Hour),
		},
	}

	handle, err := mgr.CreateHandle(pkg)
	if err != nil {
		t.Fatalf("CreateHandle failed: %v", err)
	}

	// Update handle
	handle.MarkCompleted()

	err = mgr.UpdateHandle(handle)
	if err != nil {
		t.Fatalf("UpdateHandle failed: %v", err)
	}

	// Verify update
	updatedHandle, err := mgr.GetHandle("pkg-update-001")
	if err != nil {
		t.Fatalf("GetHandle failed: %v", err)
	}

	if updatedHandle.Status != "Completed" {
		t.Errorf("Expected Status Completed, got %s", updatedHandle.Status)
	}
}

func TestExecutionManagerImpl_ListHandles(t *testing.T) {
	mgr := NewExecutionManager()

	// Create multiple handles
	for i := 0; i < 5; i++ {
		pkg := &contracts.ExecutionPackage{
			Metadata: contracts.PackageMetadata{
				ID: "pkg-list-mgr-" + string(rune('0'+i)),
			},
			Task: contracts.TaskDescriptor{
				ID: "task-list-mgr-" + string(rune('0'+i)),
			},
			Executor: contracts.ExecutorDescriptor{
				ID: "executor-list-mgr-" + string(rune('0'+i)),
			},
			Resources: contracts.ResourceAllocation{
				Provider: "claude",
				Deadline: time.Now().Add(time.Hour),
			},
		}

		_, err := mgr.CreateHandle(pkg)
		if err != nil {
			t.Fatalf("CreateHandle failed for handle %d: %v", i, err)
		}
	}

	handles := mgr.ListHandles()
	if len(handles) != 5 {
		t.Errorf("Expected 5 handles, got %d", len(handles))
	}
}

func TestExecutionManagerImpl_Cleanup(t *testing.T) {
	mgr := NewExecutionManager()

	ctx := context.Background()

	// Create handles with different states
	for i := 0; i < 3; i++ {
		pkg := &contracts.ExecutionPackage{
			Metadata: contracts.PackageMetadata{
				ID: "pkg-cleanup-" + string(rune('0'+i)),
			},
			Task: contracts.TaskDescriptor{
				ID: "task-cleanup-" + string(rune('0'+i)),
			},
			Executor: contracts.ExecutorDescriptor{
				ID: "executor-cleanup-" + string(rune('0'+i)),
			},
			Resources: contracts.ResourceAllocation{
				Provider: "claude",
				Deadline: time.Now().Add(time.Hour),
			},
		}

		handle, err := mgr.CreateHandle(pkg)
		if err != nil {
			t.Fatalf("CreateHandle failed: %v", err)
		}

		// Mark first two as completed
		if i < 2 {
			handle.MarkCompleted()
		}
	}

	// Cleanup should remove completed handles
	err = mgr.Cleanup(ctx)
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Only one handle should remain (the running one)
	handles := mgr.ListHandles()
	if len(handles) != 1 {
		t.Errorf("Expected 1 handle after cleanup, got %d", len(handles))
	}
}
