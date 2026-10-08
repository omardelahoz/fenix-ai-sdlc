package kernel

import (
	"fmt"
	"sync"

	"github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

// StateMachineImpl implements the StateMachine interface.
type StateMachineImpl struct {
	mu        sync.RWMutex
	current   contracts.KernelState
	callbacks []func(from, to contracts.KernelState)
}

// NewStateMachine creates a new StateMachine with an initial state.
func NewStateMachine(initial contracts.KernelState) *StateMachineImpl {
	return &StateMachineImpl{
		current: initial,
	}
}

// Current returns the current state.
func (sm *StateMachineImpl) Current() contracts.KernelState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.current
}

// Transition attempts to transition to a new state.
func (sm *StateMachineImpl) Transition(to contracts.KernelState) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.canTransitionInternal(sm.current, to) {
		return fmt.Errorf("invalid state transition from %s to %s", sm.current, to)
	}

	from := sm.current
	sm.current = to

	// Notify callbacks
	for _, callback := range sm.callbacks {
		callback(from, to)
	}

	return nil
}

// CanTransition checks if a transition is valid.
func (sm *StateMachineImpl) CanTransition(to contracts.KernelState) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.canTransitionInternal(sm.current, to)
}

// canTransitionInternal checks if a transition is valid (internal, without lock).
func (sm *StateMachineImpl) canTransitionInternal(from, to contracts.KernelState) bool {
	// Define valid transitions
	validTransitions := map[contracts.KernelState][]contracts.KernelState{
		contracts.StateUninitialized: {contracts.StateInitializing},
		contracts.StateInitializing:  {contracts.StateRunning, contracts.StateShutdown},
		contracts.StateRunning:       {contracts.StateSuspended, contracts.StateShuttingDown},
		contracts.StateSuspended:     {contracts.StateRunning, contracts.StateShuttingDown},
		contracts.StateShuttingDown: {contracts.StateShutdown},
		contracts.StateShutdown:      {}, // Terminal state
	}

	allowed, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowedState := range allowed {
		if allowedState == to {
			return true
		}
	}

	return false
}

// OnTransition registers a callback for state transitions.
func (sm *StateMachineImpl) OnTransition(callback func(from, to contracts.KernelState)) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.callbacks = append(sm.callbacks, callback)
}
