# SPEC-203: Core ABI (Runtime & Kernel Contracts)

## 1. Executive Summary
The most lethal anti-pattern in large Go applications is the lack of a centralized domain package, leading to circular imports.
The **Core ABI** (`/core/pkg/contracts`) solves this by defining the pure Go structs, interfaces, events, and standard errors that all Fénix components must share. It is the absolute source of truth for physical inter-module communication.

**Rule:** `pkg/contracts` cannot import ANY internal Fénix packages (`/core/internal/*`). It must be completely standalone.

## 2. Structural Layout
To prevent the `contracts` package from becoming an unmanageable monolithic file containing 150 structs, it is strictly subdivided into logical domains:

```text
/core/pkg/contracts
├── /artifacts       # Artifact Schemas, ArtifactID
├── /execution       # ExecutionPackage, ExecutionHandle, HandleID
├── /inference       # InferenceRequest, InferenceResponse, TokenBudget
├── /workflow        # WorkflowID, PipelineID, StageID
├── /types           # Shared primitive wrappers
├── /events          # Event Envelope and specific event payloads
├── /errors          # Sentinel standard errors
└── /interfaces      # Scheduler, InferenceClient, PluginClient
```

## 3. Strongly Typed IDs & Immutability
String blindness is a major cause of bugs in Go. Passing a `WorkflowID` into a function that expects a `PipelineID` because both are `string` is forbidden.
All identifiers must be strongly typed Value Objects:

```go
type ArtifactID string
type WorkflowID string
type ExecutionHandleID string
```

Furthermore, state models like `ExecutionHandle` exposed via contracts must be **Immutable** to prevent accidental mutation by observers (like the Runtime).

## 4. Layer 1: Domain Structs
Pure data transfer objects (DTOs) that pass through the system.

- **`ExecutionPackage`**: The payload sent from the Runtime to the Kernel. Contains the binary blob of instructions.
- **`ExecutionHandle`**: The atomic, read-only reference ticket returned by the Kernel.
- **`Artifact`**: The Go struct representing the JSON Canonical Schema (SPEC-110).
- **`InferenceRequest` & `InferenceResponse`**: The exact format the Processor VM uses to ask the Inference Engine for LLM evaluation.

## 5. Layer 2: Interfaces
The interfaces that allow dependency injection and modularity. Interfaces must be small and idiomatic to Go (avoiding `RuntimeManager` monoliths).

- **`KernelScheduler`**: Defines `Schedule(ExecutionPackage) -> (ExecutionHandle, error)`.
- **`InferenceClient`**: Defines `Reason(ctx, InferenceRequest) -> (InferenceResponse, error)`.
- **`PluginClient`**: The local Go wrapper around the gRPC connection to the Out-of-Process Plugin Host.

## 6. Layer 3: Events (The Envelope Pattern)
To ensure the EventBus scales cleanly, Fénix uses an `Event Envelope` pattern. This guarantees every event has standard metadata.

```go
type Event struct {
    ID        string      // Unique event ID
    Timestamp time.Time   // Emission time
    Source    string      // Component that emitted the event
    Type      string      // e.g., "ExecutionCompleted"
    Payload   any         // The specific struct payload
}
```

Standard Payloads:
- **`ExecutionCompletedPayload`**: Contains `ExecutionHandleID` and exit code.
- **`WorkflowSuspendedPayload`**: Contains `WorkflowID` and reason.

## 7. Layer 4: Standard Errors
To prevent components from inventing string-based errors, Fénix defines global Sentinel Errors in `/core/pkg/contracts/errors`.
Components use `errors.Is(err, contracts.ErrCapabilityMissing)` for deterministic error handling.

- `ErrCapabilityMissing`: Thrown when a PMF requires a capability not found in the Registry.
- `ErrInvalidArtifact`: Thrown when an artifact fails Schema Validation.
- `ErrWorkflowCancelled`: Emitted when a human or the system halts a workflow.
- `ErrExecutionTimeout`: Thrown by the Kernel Scheduler.

## 8. Separation of Concerns (Contracts vs API)
The `pkg/contracts` folder ONLY contains pure Go structs written by humans. 
It does **NOT** contain Protobuf generated code, gRPC stubs, or JSON tags specifically tuned for external network transport. 
Network-level generated code lives exclusively in `/core/pkg/api` (compiled from `/proto`). This ensures the Go domain models remain clean and agnostic to the RPC technology used.
