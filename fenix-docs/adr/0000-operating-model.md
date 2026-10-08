# ADR 0000: Fénix Operating Model

## Status
Accepted

## Context
Before designing technical components (engines, schedulers, agents), Fénix must establish a high-level operational model. The Operating Model describes the end-to-end lifecycle of a Fénix interaction—how an intent is received, translated into actionable tasks, executed, and persisted. 

Fénix is a Software Engineering Operating System. It does not simply "run a prompt"; it manages state, resolves ambiguities, coordinates parallel workstreams, and builds a comprehensive engineering knowledge graph.

## Decision
The Fénix Operating Model strictly follows this End-to-End Flow for any request:

1. **User Request (CLI/Chat):** The entry point. The user types a command like `/fenix Implement Azure AD Authentication`.
2. **Session Initialization:** Fénix Core opens a `Session`, locks the project context, and sets up tracing.
3. **Intent Resolution:** The engine analyzes the raw text to determine the system intent (e.g., `NewFeature`, `BugFix`, `Refactor`, `Discovery`).
4. **Workflow Selection:** Based on the Intent and the Project State, Fénix loads a predefined `Workflow` (e.g., `Feature Implementation Workflow`).
5. **Task Graph Generation:** The Workflow is parsed by the Execution Planner into a Directed Acyclic Graph (DAG) of `Tasks`. A Task is the atomic unit of work (e.g., `Task-1: Analyze Architecture`, `Task-2: Update DB Schema`).
6. **Parallel Scheduling:** The `Scheduler` evaluates the DAG. Tasks with no unmet dependencies are dispatched in parallel.
7. **Agent Runtime Execution:** For each dispatched Task, the `Agent Runtime` loads the assigned `Agent` (its mission and behavior), invokes the `Prompt Runtime` to build the context, and sends the payload to the `Provider` (Claude, Gemini, etc.).
8. **Validation Loop:** The AI response must fulfill the Task's `Output Contract`. The `Validation Engine` runs deterministic checks (linters, parsers, compilers). If it fails, an automated retry loop handles corrections.
9. **Artifact Production:** Validated outputs become `Artifacts`. The `Artifact Engine` writes them to the project directory.
10. **Knowledge Graph Update:** The `Engineering Knowledge Graph` detects new or modified artifacts and propagates the changes, invalidating downstream nodes that might need regeneration.
11. **State Commit & Response:** The `State Engine` commits the transaction. Fénix replies to the user, either concluding the session or escalating unresolved questions via the `Clarification Manager`.

## Consequences
- **Positive:** Agents are relegated to the bottom of the stack; they are execution nodes, not decision makers. The system is highly predictable, auditable, and resilient to LLM drift.
- **Negative:** The orchestration logic required to manage the DAG, state transitions, and validation loops is significantly complex and requires robust Go engineering.
