package kernel

import (
	"context"
	"fmt"
	"sync"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// FenixKernel is the implementation of the Kernel interface.
type FenixKernel struct {
	mu              sync.RWMutex
	state           contracts.KernelState
	executionMgr    contracts.ExecutionManager
	memoryMgr       contracts.MemoryManager
	stateMachine    contracts.StateMachine
	handles         map[string]*contracts.ExecutionHandle
	shutdownCtx     context.CancelFunc
	stateCallbacks  []func(from, to contracts.KernelState)
}

// NewKernel creates a new FenixKernel instance.
func NewKernel() *FenixKernel {
	return &FenixKernel{
		state:  contracts.StateUninitialized,
		handles: make(map[string]*contracts.ExecutionHandle),
	}
}

// Initialize initializes the Kernel and transitions to Running state.
func (k *FenixKernel) Initialize(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.state != contracts.StateUninitialized {
		return fmt.Errorf("kernel already initialized, current state: %s", k.state)
	}

	// Transition to Initializing
	k.transition(contracts.StateInitializing)

	// Initialize subsystems
	k.executionMgr = NewExecutionManager()
	k.memoryMgr = NewMemoryManager()
	k.stateMachine = NewStateMachine(k.state)

	// Transition to Running
	k.transition(contracts.StateRunning)

	return nil
}

// Shutdown gracefully shuts down the Kernel.
func (k *FenixKernel) Shutdown(ctx context.Context) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.state == contracts.StateShutdown {
		return nil
	}

	// Transition to ShuttingDown
	k.transition(contracts.StateShuttingDown)

	// Cancel all running executions
	for _, handle := range k.handles {
		if handle.IsRunning() {
			handle.Cancel()
		}
	}

	// Cleanup execution manager
	if k.executionMgr != nil {
		k.executionMgr.Cleanup(ctx)
	}

	// Transition to Shutdown
	k.transition(contracts.StateShutdown)

	return nil
}

// Execute executes an ExecutionPackage and returns an ExecutionHandle.
func (k *FenixKernel) Execute(ctx context.Context, pkg *contracts.ExecutionPackage) (*contracts.ExecutionHandle, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	if k.state != contracts.StateRunning {
		return nil, fmt.Errorf("kernel not in running state, current state: %s", k.state)
	}

	// Validate the package
	if err := pkg.Validate(); err != nil {
		return nil, fmt.Errorf("package validation failed: %w", err)
	}

	// Create execution handle
	handle, err := k.executionMgr.CreateHandle(pkg)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution handle: %w", err)
	}

	// Store handle
	k.handles[pkg.Metadata.ID] = handle

	// TODO: Actually execute the task (this would involve the Scheduler and Execution Engine)
	// For now, we'll just mark it as completed for demonstration
	handle.MarkCompleted()

	return handle, nil
}

// Cancel cancels an execution by package ID.
func (k *FenixKernel) Cancel(ctx context.Context, packageID string) error {
	k.mu.Lock()
	defer k.mu.Unlock()

	handle, exists := k.handles[packageID]
	if !exists {
		return fmt.Errorf("execution handle not found for package ID: %s", packageID)
	}

	handle.Cancel()
	return nil
}

// GetHandle retrieves an ExecutionHandle by package ID.
func (k *FenixKernel) GetHandle(packageID string) (*contracts.ExecutionHandle, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()

	handle, exists := k.handles[packageID]
	if !exists {
		return nil, fmt.Errorf("execution handle not found for package ID: %s", packageID)
	}

	return handle, nil
}

// ListHandles returns all active ExecutionHandles.
func (k *FenixKernel) ListHandles() []*contracts.ExecutionHandle {
	k.mu.RLock()
	defer k.mu.RUnlock()

	handles := make([]*contracts.ExecutionHandle, 0, len(k.handles))
	for _, handle := range k.handles {
		handles = append(handles, handle)
	}

	return handles
}

// State returns the current Kernel state.
func (k *FenixKernel) State() contracts.KernelState {
	k.mu.RLock()
	defer k.mu.RUnlock()

	return k.state
}

// transition performs a state transition and notifies callbacks.
func (k *FenixKernel) transition(to contracts.KernelState) {
	from := k.state
	k.state = to

	// Notify callbacks
	for _, callback := range k.stateCallbacks {
		callback(from, to)
	}
}

// OnTransition registers a callback for state transitions.
func (k *FenixKernel) OnTransition(callback func(from, to contracts.KernelState)) {
	k.mu.Lock()
	defer k.mu.Unlock()

	k.stateCallbacks = append(k.stateCallbacks, callback)
}
