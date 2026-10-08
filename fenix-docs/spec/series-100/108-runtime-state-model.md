# SPEC-108: Runtime State Model (Control Plane)

## 1. Executive Summary
The most critical boundary in the Fénix Operating System is the separation between execution and orchestration. 

**The Core Axiom: The Kernel owns execution state. The Runtime owns orchestration state.**

The Runtime must never duplicate the Kernel's state (e.g., tracking whether a thread is `Executing` or `Faulted`). Instead, the Runtime acts as the **Control Plane** (akin to Kubernetes), maintaining an orchestration tree constructed entirely via Event Sourcing. It projects Kernel events onto semantic, high-level states (e.g., `WaitingJoin`, `Resolving DAG`).

## 2. The Orchestration Tree
The Runtime models a software project as a hierarchy of instances. Each node possesses an ID, pointers to parents/children, and references to the specific Kernel artifacts it manages.

```text
WorkflowInstance (e.g., "SaaS v1.0 Launch")
    │
    ├── PipelineInstance (e.g., "Main Development")
    │      │
    │      ├── StageInstance (e.g., "Backend & Frontend Dev")
    │      │      │
    │      │      ├── ProcessorInstance (e.g., "Go Backend")
    │      │      │      │
    │      │      │      └── ExecutionHandle (Kernel ID: 512)
    │      │      │
    │      │      └── ProcessorInstance (e.g., "React Frontend")
    │      │             │
    │      │             └── ExecutionHandle (Kernel ID: 513)
```

## 3. Orchestration Finite State Machines (FSM)
Because the Runtime does not manage CPU or Memory, its State Machines are purely logical and semantic.

### Workflow FSM
Governs the macroscopic lifecycle of an SDLC process.
`Created` -> `WaitingTrigger` (Listening for events) -> `Running` -> `Paused` (Global suspension) -> `Completed` / `Archived`

### Pipeline FSM
Governs the DAG evaluation.
`Created` -> `Resolving DAG` (Static Analysis) -> `Running` -> `Blocked` (Awaiting dependencies) -> `Completed` / `Cancelled`

### Stage FSM
Governs concurrency and fan-out/fan-in boundaries.
`Pending` -> `FanOut` (Spawning processors) -> `WaitingJoin` (Applying Quorum/Wait-All logic) -> `Gate` (Awaiting human approval) -> `Completed`

### Processor FSM
Governs the VM initialization. Note the absence of a `Running` state; that belongs to the Kernel.
`Ready` -> `WaitingExecution` -> `Attached` (Bound to a Kernel Handle) -> `Finished`

## 4. The Event Projection Layer
The Runtime does not poll the Kernel. It subscribes to the `Kernel EventBus`.

When the Kernel emits an `ExecutionCompleted(Handle=512)` event, the Runtime's **Projection Engine** intercepts it. The Engine finds the `ProcessorInstance` bound to `Handle 512`, transitions it to `Finished`, and triggers an evaluation of the parent `StageInstance`. If the Stage was in `WaitingJoin` and the Quorum is now met, the Stage transitions to `Completed`, recursively updating the tree.

## 5. Runtime Journal & Tiered Snapshots
The Kernel maintains a strict Journal of system calls, memory allocations, and VFS changes.
The Runtime maintains its own isolated **Orchestration Journal**, logging structural changes to the DAG and human approvals.

Snapshots are tiered. A `Workflow Snapshot` does not duplicate the gigabytes of Kernel memory; it merely serializes the structural tree of IDs. This guarantees that if the Fénix daemon is forcefully restarted, the Runtime reads the orchestration snapshot, queries the Kernel for the status of the `ExecutionHandles`, and instantly resynchronizes its projection tree.

## 6. Observability: Status Projection
Because the state is highly structured and segregated, querying the system state (`fenix status`) is extremely lightweight and deterministic:

```text
Workflow: StandardSaaS (Running)
│
├── Pipeline: Discovery (✔)
├── Pipeline: Development (Running)
│   ├── Stage: Database (✔)
│   ├── Stage: Backend (Attached -> Kernel: Executing)
│   └── Stage: Frontend (Gate -> Awaiting SolutionArchitect)
└── Pipeline: Release (Pending)
```

## 7. Conclusion
By treating the Runtime as a pure projection layer over the immutable execution state of the Kernel, Fénix completely eliminates state-bifurcation bugs. The architecture maps perfectly to modern distributed systems, giving architects precise, high-level control over multi-week AI workflows without worrying about low-level thread management.
