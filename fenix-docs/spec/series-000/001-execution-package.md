# SPEC-001: The Execution Package (PCB)

## 1. Executive Summary
Fénix is an **AI-Native Engineering Kernel** (a Microkernel architecture). In this operating system, the Kernel schedules and manages units of work using a Process Control Block (PCB). 

The **Execution Package** is the immutable PCB of the Fénix Kernel. It is a strictly typed internal struct that travels through the Kernel (Scheduler, Memory Manager, Execution Manager). 

**Critical Rule:** The `ExecutionPackage` is completely immutable. Any transition or change in state generates a new version (like a Git commit). Mutable runtime tracking is handled separately by an `ExecutionHandle`.

## 2. The Internal Data Structure
The Fénix Kernel is implemented in Go. The PCB is decoupled from business logic; it does not know what an "Agent" is, it only knows about `ExecutorDescriptors`.

### 2.1 The Immutable Execution Package (PCB)
```go
package kernel

import "time"

// ExecutionPackage is the immutable Process Control Block (PCB) of the Fénix Kernel.
type ExecutionPackage struct {
    Metadata  PackageMetadata
    Task      TaskDescriptor
    Executor  ExecutorDescriptor  // Replaces "Agent". The Kernel is agnostic to the actor type.
    Snapshot  ExecutionSnapshot   // The frozen state (L0 Cache)
    Policies  ExecutionPolicies
    Resources ResourceAllocation  // Cost, Tokens, Priority
    Contracts Contracts
}

type PackageMetadata struct {
    Id            string
    Version       string // PCB Versioning (e.g., v1, v2) for time-travel and auditing
    SchemaVersion string // Version of this struct schema
    KernelVersion string // Version of the Fénix Kernel that created this PCB
    TraceId       string
    CorrelationId string
}

type TaskDescriptor struct {
    Id   string
    Type string 
}

// ExecutorDescriptor abstracts Agents, Skills, Plugins, or Workflow Steps.
type ExecutorDescriptor struct {
    Id           string
    Type         string // "Agent", "Skill", "Plugin"
    Version      string
    Capabilities []string
}

// ExecutionSnapshot is strictly the frozen state of the repo. (No errors or state).
type ExecutionSnapshot struct {
    AstNodes         []string
    MemoryReferences []string
}

type ExecutionPolicies struct {
    MaxTokens           int
    MaxFiles            int
    MaxAstDepth         int
    CompressionStrategy string
    RetryPolicy         RetryConfig // Kernel-managed retries
}

type RetryConfig struct {
    MaxRetries      int
    BackoffStrategy string
}

// ResourceAllocation tracks budget and deadlines. (Decided by the Execution Planner).
type ResourceAllocation struct {
    Provider        string // e.g., "claude-3.5-sonnet"
    EstimatedTokens int
    ReservedTokens  int
    EstimatedCost   float64
    Priority        int
    Deadline        time.Time
}

type Contracts struct {
    ExpectedFormat  string
    JsonSchema      string
    ValidationRules []string
}
```

### 2.2 The Mutable Execution Handle
Because the `ExecutionPackage` is immutable, the Kernel tracks live execution state using an `ExecutionHandle`. This lives in RAM and is owned by the `ExecutionManager`.

```go
package kernel

import "time"

// ExecutionHandle tracks the live, mutable state of a running ExecutionPackage.
type ExecutionHandle struct {
    PackageId         string
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
```

## 3. The Lifecycle of the Execution Package (PCB)

### Phase 1: Queuing the Execution Descriptor (Lazy Evaluation)
The `Workflow Engine` generates tasks. However, building an `ExecutionPackage` (specifically the `ExecutionSnapshot`) is expensive in RAM and CPU. 
Therefore, the `Task Scheduler` does **not** receive full Packages. It receives an `ExecutionDescriptor` (a lightweight pointer).

### Phase 2: Materialization (Worker Available)
Only when a Worker thread becomes free does the Kernel pop the `ExecutionDescriptor` and hand it to the `Memory Manager`.
1. **Discover:** The Manager resolves dependencies.
2. **Snapshot Build:** The `ExecutionSnapshot` is built via L3 Cache mapping.
3. **Immutability:** The full `ExecutionPackage` (v1) is stamped and frozen.

### Phase 3: Execution & Serialization
The Worker thread reads the immutable PCB. The `Prompt Runtime` serializes the struct into a Provider-specific API payload. The `Provider Adapter` (Device Driver) handles the actual HTTP call. The `ExecutionHandle` tracks latency and status.

### Phase 4: Validation & Versioning
The LLM response is validated by the `Validation Manager`.
- **Success:** The `ExecutionHandle` is marked `Completed`. The Artifacts are saved to the `Artifact Store` (File System).
- **Failure:** The error is appended to the `ExecutionHandle.Errors`. If a retry is triggered, a **new** immutable PCB is generated (`Version: v2`) with an updated Snapshot, preserving the audit trail of the failure.

## 4. Conclusion
By making the PCB completely immutable and decoupling it from both business logic (Agents) and mutable state (ExecutionHandle), Fénix achieves deterministic execution. It natively supports time-travel debugging, extreme RAM efficiency through Lazy Evaluation, and a true Microkernel separation of concerns.
