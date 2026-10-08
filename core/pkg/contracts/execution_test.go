package contracts

import (
	"testing"
	"time"
)

func TestExecutionPackage_Validate(t *testing.T) {
	tests := []struct {
		name    string
		pkg     *ExecutionPackage
		wantErr bool
	}{
		{
			name: "valid package",
			pkg: &ExecutionPackage{
				Metadata: PackageMetadata{
					ID: "pkg-123",
				},
				Task: TaskDescriptor{
					ID: "task-456",
				},
				Executor: ExecutorDescriptor{
					ID: "executor-789",
				},
				Resources: ResourceAllocation{
					Provider: "claude-3.5-sonnet",
					Deadline: time.Now().Add(time.Hour),
				},
			},
			wantErr: false,
		},
		{
			name: "missing metadata ID",
			pkg: &ExecutionPackage{
				Metadata: PackageMetadata{},
				Task: TaskDescriptor{
					ID: "task-456",
				},
				Executor: ExecutorDescriptor{
					ID: "executor-789",
				},
				Resources: ResourceAllocation{
					Provider: "claude-3.5-sonnet",
					Deadline: time.Now().Add(time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "missing task ID",
			pkg: &ExecutionPackage{
				Metadata: PackageMetadata{
					ID: "pkg-123",
				},
				Task: TaskDescriptor{},
				Executor: ExecutorDescriptor{
					ID: "executor-789",
				},
				Resources: ResourceAllocation{
					Provider: "claude-3.5-sonnet",
					Deadline: time.Now().Add(time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "missing executor ID",
			pkg: &ExecutionPackage{
				Metadata: PackageMetadata{
					ID: "pkg-123",
				},
				Task: TaskDescriptor{
					ID: "task-456",
				},
				Executor: ExecutorDescriptor{},
				Resources: ResourceAllocation{
					Provider: "claude-3.5-sonnet",
					Deadline: time.Now().Add(time.Hour),
				},
			},
			wantErr: true,
		},
		{
			name: "missing deadline",
			pkg: &ExecutionPackage{
				Metadata: PackageMetadata{
					ID: "pkg-123",
				},
				Task: TaskDescriptor{
					ID: "task-456",
				},
				Executor: ExecutorDescriptor{
					ID: "executor-789",
				},
				Resources: ResourceAllocation{
					Provider: "claude-3.5-sonnet",
					Deadline: time.Time{},
				},
			},
			wantErr: true,
		},
		{
			name: "missing provider",
			pkg: &ExecutionPackage{
				Metadata: PackageMetadata{
					ID: "pkg-123",
				},
				Task: TaskDescriptor{
					ID: "task-456",
				},
				Executor: ExecutorDescriptor{
					ID: "executor-789",
				},
				Resources: ResourceAllocation{
					Provider: "",
					Deadline: time.Now().Add(time.Hour),
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pkg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecutionPackage.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutionPackage_NewVersion(t *testing.T) {
	original := &ExecutionPackage{
		Metadata: PackageMetadata{
			ID:      "pkg-123",
			Version: "v1",
		},
		Task: TaskDescriptor{
			ID: "task-456",
		},
		Executor: ExecutorDescriptor{
			ID: "executor-789",
		},
	}

	newPkg := original.NewVersion(func(pkg *ExecutionPackage) {
		pkg.Task.ID = "task-updated"
	})

	// Verify original is unchanged
	if original.Task.ID != "task-456" {
		t.Errorf("Original package was modified")
	}

	// Verify new package has updated field
	if newPkg.Task.ID != "task-updated" {
		t.Errorf("New package Task.ID = %v, want task-updated", newPkg.Task.ID)
	}

	// Verify version incremented
	if newPkg.Metadata.Version != "v2" {
		t.Errorf("New package version = %v, want v2", newPkg.Metadata.Version)
	}

	// Verify ID is preserved
	if newPkg.Metadata.ID != "pkg-123" {
		t.Errorf("New package ID = %v, want pkg-123", newPkg.Metadata.ID)
	}
}

func TestExecutionPackage_NewVersion_EmptyVersion(t *testing.T) {
	original := &ExecutionPackage{
		Metadata: PackageMetadata{
			ID:      "pkg-123",
			Version: "",
		},
	}

	newPkg := original.NewVersion(func(pkg *ExecutionPackage) {})

	if newPkg.Metadata.Version != "v1" {
		t.Errorf("Empty version should become v1, got %v", newPkg.Metadata.Version)
	}
}

func TestExecutionHandle_NewExecutionHandle(t *testing.T) {
	pkg := &ExecutionPackage{
		Metadata: PackageMetadata{
			ID:      "pkg-123",
			Version: "v1",
		},
	}

	handle := NewExecutionHandle(pkg)

	if handle.PackageID != "pkg-123" {
		t.Errorf("PackageID = %v, want pkg-123", handle.PackageID)
	}

	if handle.CurrentVersion != "v1" {
		t.Errorf("CurrentVersion = %v, want v1", handle.CurrentVersion)
	}

	if handle.Attempt != 1 {
		t.Errorf("Attempt = %v, want 1", handle.Attempt)
	}

	if handle.Status != "Running" {
		t.Errorf("Status = %v, want Running", handle.Status)
	}

	if handle.CurrentPhase != "Pending" {
		t.Errorf("CurrentPhase = %v, want Pending", handle.CurrentPhase)
	}
}

func TestExecutionHandle_MarkCompleted(t *testing.T) {
	pkg := &ExecutionPackage{
		Metadata: PackageMetadata{ID: "pkg-123"},
	}
	handle := NewExecutionHandle(pkg)

	handle.MarkCompleted()

	if !handle.IsCompleted() {
		t.Errorf("Expected IsCompleted to be true")
	}

	if handle.Status != "Completed" {
		t.Errorf("Status = %v, want Completed", handle.Status)
	}

	if handle.EndedAt.IsZero() {
		t.Errorf("EndedAt should be set")
	}

	if handle.Duration == 0 {
		t.Errorf("Duration should be calculated")
	}
}

func TestExecutionHandle_MarkFaulted(t *testing.T) {
	pkg := &ExecutionPackage{
		Metadata: PackageMetadata{ID: "pkg-123"},
	}
	handle := NewExecutionHandle(pkg)

	err := &ValidationError{Field: "test", Message: "test error"}
	handle.MarkFaulted(err)

	if !handle.IsFaulted() {
		t.Errorf("Expected IsFaulted to be true")
	}

	if handle.Status != "Faulted" {
		t.Errorf("Status = %v, want Faulted", handle.Status)
	}

	if len(handle.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(handle.Errors))
	}
}

func TestExecutionHandle_Cancel(t *testing.T) {
	pkg := &ExecutionPackage{
		Metadata: PackageMetadata{ID: "pkg-123"},
	}
	handle := NewExecutionHandle(pkg)

	// Verify channel is open initially
	select {
	case <-handle.CancellationToken:
		t.Errorf("Channel should be open initially")
	default:
	}

	handle.Cancel()

	// Verify channel is closed after cancel
	select {
	case <-handle.CancellationToken:
		// Channel is closed, as expected
	default:
		t.Errorf("Channel should be closed after cancel")
	}
}

func TestExecutionHandle_IncrementAttempt(t *testing.T) {
	pkg := &ExecutionPackage{
		Metadata: PackageMetadata{ID: "pkg-123"},
	}
	handle := NewExecutionHandle(pkg)

	if handle.Attempt != 1 {
		t.Errorf("Initial attempt = %v, want 1", handle.Attempt)
	}

	handle.IncrementAttempt()

	if handle.Attempt != 2 {
		t.Errorf("Attempt after increment = %v, want 2", handle.Attempt)
	}

	if handle.Status != "Running" {
		t.Errorf("Status after increment = %v, want Running", handle.Status)
	}
}
