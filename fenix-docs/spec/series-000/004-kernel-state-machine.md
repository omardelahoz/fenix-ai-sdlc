# SPEC-004: Kernel State Machine

## 1. Executive Summary
In complex distributed systems, tracking the lifecycle of an execution unit using loose string updates (e.g., `status = "Running"`) inevitably leads to race conditions, orphaned processes, and inconsistent states.

The Fénix Kernel enforces absolute determinism through the **Kernel State Machine**. This Finite State Machine (FSM) strictly dictates the valid lifecycle of an `ExecutionHandle`. No component in the system can arbitrarily change the state of a process; they must submit a state transition request to the Kernel, which evaluates it against the FSM matrix. Attempting an invalid transition triggers a `Kernel Panic` for that specific task.

## 2. The Core States
The lifecycle of an execution unit is tracked in the mutable `ExecutionHandle` across 8 primary states:

1. **Created:** The `ExecutionDescriptor` is born in the Execution Planner.
2. **Queued:** The Descriptor is submitted to the `Scheduler`.
3. **Scheduled:** The `Dispatcher` assigns the Descriptor to a specific Worker thread based on NUMA/Affinity.
4. **Materialized:** The `MemoryManager` has successfully built the immutable `ExecutionPackage` (PCB) and attached it to the Handle.
5. **Executing:** The `ExecutionManager` has invoked the `AIDriver`, and the Provider is actively streaming/processing.
6. **Validating:** The Driver has returned a response, and the `ValidationManager` is evaluating the Output Contracts.
7. **Persisting:** Validation passed. The `VFS` (Virtual File System) is writing the artifacts and committing the snapshot.
8. **Completed:** The final, successful terminal state.

## 3. Interrupt and Fault States
Execution rarely follows the happy path perfectly. The FSM defines 4 secondary states to handle exceptions:

1. **Suspended:** Execution paused (e.g., waiting for human clarification or hitting a rate-limit backoff).
2. **Cancelled:** The Kernel forcibly terminated the task via the `CancellationToken` (e.g., due to cache invalidation or manual abort).
3. **Retrying:** A Validation fault occurred, but the `Policies.RetryPolicy` allows another attempt.
4. **Faulted (Terminal):** An unrecoverable error (e.g., OOM, max retries exceeded, malformed output).

## 4. Valid Transitions Matrix
The Kernel strictly enforces these transitions. Any transition not explicitly defined here is illegal.

### The Happy Path
- `Created` -> `Queued`
- `Queued` -> `Scheduled`
- `Scheduled` -> `Materialized`
- `Materialized` -> `Executing`
- `Executing` -> `Validating`
- `Validating` -> `Persisting`
- `Persisting` -> `Completed`

### Interruptions and Recovery
- `Executing` -> `Suspended` (Waiting for Input/Rate Limit)
- `Suspended` -> `Executing` (Resumed)
- `Executing` -> `Cancelled` (Cache invalidation / Timeout)
- `Queued` | `Scheduled` | `Materialized` -> `Cancelled` (Aborted before execution)
- `Validating` -> `Retrying` (Contract violation, retry allowed)
- `Retrying` -> `Materialized` (The MemoryManager must build a *new* PCB version for the retry)

### Terminal Faults
- `Validating` -> `Faulted` (Max retries exceeded)
- `Materialized` -> `Faulted` (OOM or dependency resolution failure)
- `Executing` -> `Faulted` (Driver catastrophic failure, e.g., Provider down)

## 5. Implementation (The State Machine Engine)
The FSM is implemented in Go within the `ExecutionManager`. It utilizes a Mutex-locked transition function.

```go
package kernel

import (
    "fmt"
    "sync"
)

type State string

const (
    StateCreated      State = "Created"
    StateQueued       State = "Queued"
    StateScheduled    State = "Scheduled"
    StateMaterialized State = "Materialized"
    StateExecuting    State = "Executing"
    StateValidating   State = "Validating"
    StatePersisting   State = "Persisting"
    StateCompleted    State = "Completed"
    StateSuspended    State = "Suspended"
    StateCancelled    State = "Cancelled"
    StateRetrying     State = "Retrying"
    StateFaulted      State = "Faulted"
)

// ValidTransitions defines the strict matrix of allowed state jumps.
var ValidTransitions = map[State][]State{
    StateCreated:      {StateQueued, StateCancelled},
    StateQueued:       {StateScheduled, StateCancelled},
    StateScheduled:    {StateMaterialized, StateCancelled},
    StateMaterialized: {StateExecuting, StateFaulted, StateCancelled},
    StateExecuting:    {StateValidating, StateSuspended, StateCancelled, StateFaulted},
    StateSuspended:    {StateExecuting, StateCancelled},
    StateValidating:   {StatePersisting, StateRetrying, StateFaulted},
    StateRetrying:     {StateMaterialized, StateCancelled},
    StatePersisting:   {StateCompleted, StateFaulted},
    StateCompleted:    {}, // Terminal
    StateCancelled:    {}, // Terminal
    StateFaulted:      {}, // Terminal
}

type ExecutionHandle struct {
    mu      sync.RWMutex
    state   State
    // ... other fields
}

// Transition attempts to safely move the Handle to a new state.
func (h *ExecutionHandle) Transition(next State) error {
    h.mu.Lock()
    defer h.mu.Unlock()

    allowed := false
    for _, valid := range ValidTransitions[h.state] {
        if next == valid {
            allowed = true
            break
        }
    }

    if !allowed {
        // Triggers a localized Kernel Panic for this execution unit.
        return fmt.Errorf("KERNEL PANIC: Invalid state transition from %s to %s", h.state, next)
    }

    h.state = next
    return nil
}
```

## 6. Conclusion
By strictly codifying the state transitions, the Fénix Kernel eliminates phantom bugs related to concurrency. A Worker thread cannot mistakenly execute a Cancelled task, and a validation failure correctly routes back through Materialization to generate a fresh, immutable PCB (v2) before Retrying. This FSM is the bedrock of Fénix's execution stability.
