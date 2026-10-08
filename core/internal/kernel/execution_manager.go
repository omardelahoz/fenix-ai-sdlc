package kernel

import (
	"context"
	"fmt"
	"sync"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// ExecutionManagerImpl implements the ExecutionManager interface.
type ExecutionManagerImpl struct {
	mu      sync.RWMutex
	handles map[string]*contracts.ExecutionHandle
}

// NewExecutionManager creates a new ExecutionManager.
func NewExecutionManager() *ExecutionManagerImpl {
	return &ExecutionManagerImpl{
		handles: make(map[string]*contracts.ExecutionHandle),
	}
}

// CreateHandle creates a new ExecutionHandle for an ExecutionPackage.
func (m *ExecutionManagerImpl) CreateHandle(pkg *contracts.ExecutionPackage) (*contracts.ExecutionHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if handle already exists
	if _, exists := m.handles[pkg.Metadata.ID]; exists {
		return nil, fmt.Errorf("execution handle already exists for package ID: %s", pkg.Metadata.ID)
	}

	// Create new handle
	handle := contracts.NewExecutionHandle(pkg)
	m.handles[pkg.Metadata.ID] = handle

	return handle, nil
}

// UpdateHandle updates an existing ExecutionHandle.
func (m *ExecutionManagerImpl) UpdateHandle(handle *contracts.ExecutionHandle) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.handles[handle.PackageID]; !exists {
		return fmt.Errorf("execution handle not found for package ID: %s", handle.PackageID)
	}

	m.handles[handle.PackageID] = handle
	return nil
}

// GetHandle retrieves an ExecutionHandle by ID.
func (m *ExecutionManagerImpl) GetHandle(packageID string) (*contracts.ExecutionHandle, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	handle, exists := m.handles[packageID]
	if !exists {
		return nil, fmt.Errorf("execution handle not found for package ID: %s", packageID)
	}

	return handle, nil
}

// ListHandles returns all active ExecutionHandles.
func (m *ExecutionManagerImpl) ListHandles() []*contracts.ExecutionHandle {
	m.mu.RLock()
	defer m.mu.RUnlock()

	handles := make([]*contracts.ExecutionHandle, 0, len(m.handles))
	for _, handle := range m.handles {
		handles = append(handles, handle)
	}

	return handles
}

// Cleanup removes completed/faulted handles.
func (m *ExecutionManagerImpl) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, handle := range m.handles {
		if handle.IsCompleted() || handle.IsFaulted() {
			delete(m.handles, id)
		}
	}

	return nil
}
