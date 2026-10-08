package contracts

import "context"

// Kernel is the main interface for the Fénix Kernel.
// It manages ExecutionPackages, schedules tasks, and coordinates between subsystems.
type Kernel interface {
	// Lifecycle
	Initialize(ctx context.Context) error
	Shutdown(ctx context.Context) error

	// Execution Management
	Execute(ctx context.Context, pkg *ExecutionPackage) (*ExecutionHandle, error)
	Cancel(ctx context.Context, packageID string) error

	// Query
	GetHandle(packageID string) (*ExecutionHandle, error)
	ListHandles() []*ExecutionHandle

	// State
	State() KernelState
}

// KernelState represents the current state of the Kernel.
type KernelState string

const (
	StateUninitialized KernelState = "Uninitialized"
	StateInitializing  KernelState = "Initializing"
	StateRunning       KernelState = "Running"
	StateSuspended     KernelState = "Suspended"
	StateShuttingDown KernelState = "ShuttingDown"
	StateShutdown      KernelState = "Shutdown"
)

// MemoryManager manages the memory references and L0 cache for ExecutionPackages.
type MemoryManager interface {
	// Resolve resolves memory references in an ExecutionSnapshot
	Resolve(ctx context.Context, snapshot *ExecutionSnapshot) ([]string, error)

	// Cache caches an ExecutionSnapshot for fast lookup
	Cache(ctx context.Context, snapshotID string, snapshot *ExecutionSnapshot) error

	// Invalidate invalidates cached snapshots
	Invalidate(ctx context.Context, snapshotID string) error

	// Get retrieves a cached snapshot
	Get(ctx context.Context, snapshotID string) (*ExecutionSnapshot, error)
}

// ExecutionManager manages the lifecycle of ExecutionHandles.
type ExecutionManager interface {
	// CreateHandle creates a new ExecutionHandle for an ExecutionPackage
	CreateHandle(pkg *ExecutionPackage) (*ExecutionHandle, error)

	// UpdateHandle updates an existing ExecutionHandle
	UpdateHandle(handle *ExecutionHandle) error

	// GetHandle retrieves an ExecutionHandle by ID
	GetHandle(packageID string) (*ExecutionHandle, error)

	// ListHandles returns all active ExecutionHandles
	ListHandles() []*ExecutionHandle

	// Cleanup removes completed/faulted handles
	Cleanup(ctx context.Context) error
}

// StateMachine manages the Kernel state transitions.
type StateMachine interface {
	// Current returns the current state
	Current() KernelState

	// Transition attempts to transition to a new state
	Transition(to KernelState) error

	// CanTransition checks if a transition is valid
	CanTransition(to KernelState) bool

	// OnTransition registers a callback for state transitions
	OnTransition(callback func(from, to KernelState))
}

// ContextModel represents the execution context for tasks.
type ContextModel interface {
	// WithContext creates a new context with the given parent
	WithContext(parent context.Context) context.Context

	// WithValue adds a value to the context
	WithValue(key, value interface{}) ContextModel

	// GetValue retrieves a value from the context
	GetValue(key interface{}) interface{}

	// TraceID returns the trace ID for the context
	TraceID() string

	// CorrelationID returns the correlation ID for the context
	CorrelationID() string
}
