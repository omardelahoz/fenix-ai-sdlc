package contracts

import "time"

// ExecutionHandle tracks the live, mutable state of a running ExecutionPackage.
// This lives in RAM and is owned by the ExecutionManager.
// The ExecutionPackage is immutable; the ExecutionHandle tracks the runtime state.
type ExecutionHandle struct {
	PackageID         string
	CurrentVersion    string
	Attempt           int
	StartedAt         time.Time
	EndedAt           time.Time
	Duration          time.Duration
	CurrentPhase      string // Pending, Discover, Executing, Validating, Completed
	Status            string // Running, Suspended, Faulted
	WorkerNode        string
	ProviderLatency   time.Duration
	CacheHit          bool
	Errors            []error // Runtime faults, validation failures
	CancellationToken <-chan struct{}
}

// NewExecutionHandle creates a new ExecutionHandle for a given ExecutionPackage.
func NewExecutionHandle(pkg *ExecutionPackage) *ExecutionHandle {
	return &ExecutionHandle{
		PackageID:         pkg.Metadata.ID,
		CurrentVersion:    pkg.Metadata.Version,
		Attempt:           1,
		StartedAt:         time.Now(),
		CurrentPhase:      "Pending",
		Status:            "Running",
		CancellationToken: make(chan struct{}),
	}
}

// MarkCompleted marks the execution as completed.
func (h *ExecutionHandle) MarkCompleted() {
	h.Status = "Completed"
	h.CurrentPhase = "Completed"
	h.EndedAt = time.Now()
	h.Duration = h.EndedAt.Sub(h.StartedAt)
	close(h.CancellationToken)
}

// MarkFaulted marks the execution as faulted with an error.
func (h *ExecutionHandle) MarkFaulted(err error) {
	h.Status = "Faulted"
	h.CurrentPhase = "Faulted"
	h.Errors = append(h.Errors, err)
	h.EndedAt = time.Now()
	h.Duration = h.EndedAt.Sub(h.StartAt)
	close(h.CancellationToken)
}

// MarkSuspended marks the execution as suspended.
func (h *ExecutionHandle) MarkSuspended() {
	h.Status = "Suspended"
}

// MarkResumed marks the execution as resumed from suspension.
func (h *ExecutionHandle) MarkResumed() {
	h.Status = "Running"
}

// Cancel cancels the execution by sending to the cancellation channel.
func (h *ExecutionHandle) Cancel() {
	select {
	case <-h.CancellationToken:
		// Already closed
	default:
		close(h.CancellationToken)
	}
}

// IncrementAttempt increments the attempt counter for retries.
func (h *ExecutionHandle) IncrementAttempt() {
	h.Attempt++
	h.StartedAt = time.Now()
	h.Status = "Running"
	h.CurrentPhase = "Executing"
	h.CancellationToken = make(chan struct{})
}

// SetPhase updates the current phase of execution.
func (h *ExecutionHandle) SetPhase(phase string) {
	h.CurrentPhase = phase
}

// AddError adds an error to the handle.
func (h *ExecutionHandle) AddError(err error) {
	h.Errors = append(h.Errors, err)
}

// IsCompleted returns true if the execution is completed.
func (h *ExecutionHandle) IsCompleted() bool {
	return h.Status == "Completed"
}

// IsFaulted returns true if the execution is faulted.
func (h *ExecutionHandle) IsFaulted() bool {
	return h.Status == "Faulted"
}

// IsRunning returns true if the execution is currently running.
func (h *ExecutionHandle) IsRunning() bool {
	return h.Status == "Running"
}

// IsSuspended returns true if the execution is suspended.
func (h *ExecutionHandle) IsSuspended() bool {
	return h.Status == "Suspended"
}
