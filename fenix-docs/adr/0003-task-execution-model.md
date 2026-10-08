# ADR 0003: Task Execution Model

## Status
Accepted

## Context
Previous architectures placed the "Agent" at the center of the SDLC. This led to non-deterministic loops where agents decided what to do next. To build a true operating system, the center of execution must be the `Task`.

## Decision
A `Task` is the fundamental unit of work in Fénix.

### The Task Contract
Every Task is instantiated with a strict schema:
- **ID:** Unique identifier (e.g., `TASK-402`).
- **Workflow / Phase:** The parent process.
- **Agent:** The assigned persona.
- **Inputs:** The specific Context payload required.
- **Outputs:** The exact JSON Schema or file structure expected.
- **Dependencies:** Other Task IDs that must succeed before this can run.
- **Priority:** Execution weight.
- **Retries:** Remaining retry attempts.
- **Status:** `Pending`, `Running`, `WaitingForClarification`, `Failed`, `Success`.

### The Lifecycle
1. **Creation:** A Workflow generates a DAG of Tasks.
2. **Scheduling:** The Scheduler polls for `Pending` Tasks with zero unmet dependencies.
3. **Execution:** The Task is dispatched. The `Prompt Runtime` builds the payload, and the `Agent Runtime` communicates with the Provider.
4. **Validation:** The Provider returns the Output. The `Validation Engine` asserts the Output against the Task's Output Contract.
5. **Resolution:** If valid, the Task is marked `Success`. The Scheduler unlocks dependent Tasks. If invalid, the `Retries` counter decrements, and the Task is re-queued with the validation error injected as feedback.

## Consequences
- **Positive:** Absolute determinism. Fénix can pause, serialize, and resume a DAG of tasks effortlessly. We can trace exactly why a feature took 10 minutes (by summing task execution times).
- **Negative:** Designing Workflows is harder because it requires explicitly defining the Inputs and Outputs for every possible step in the SDLC.
