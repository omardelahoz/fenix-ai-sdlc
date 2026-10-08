# SPEC-206: CLI Architecture (Control Plane Interface)

## 1. Executive Summary
Fénix is not a static script that executes a workflow and exits. Fénix is a persistent Operating System designed to manage long-running inference sessions, event buses, and distributed plugin architectures.
Therefore, the CLI is merely a **Thin Client**. It does not execute workflows. It communicates with a long-running **Fénix Daemon** via local Inter-Process Communication (IPC).

## 2. The Client-Server Paradigm
The Fénix CLI follows the architectural pattern of Docker and the VS Code Language Server Protocol (LSP).

```text
                +----------------------+
                |     fenix CLI        |
                |   (Thin Client)      |
                +----------+-----------+
                           |
                    Local gRPC / UDS
                           |
                 +---------v---------+
                 |   Fénix Daemon    |
                 |-------------------|
                 | Runtime           |
                 | Kernel            |
                 | Inference Engine  |
                 | Plugin Host       |
                 +---------+---------+
                           |
                     Plugins (gRPC)
```

The CLI executable contains **no business logic**. It does not import `/core/internal/kernel` or `/core/internal/runtime`. It only imports `/core/pkg/api` (the gRPC Protobuf definitions) and `/core/pkg/contracts`.

## 3. Local IPC (Unix Domain Sockets & Named Pipes)
To ensure maximum security and avoid port-binding conflicts (e.g., `localhost:8080` already in use), Fénix uses native OS inter-process communication:
- **Linux / macOS**: Unix Domain Sockets (UDS). e.g., `/tmp/fenix.sock`.
- **Windows**: Named Pipes. e.g., `\\.\pipe\fenix_daemon`.

There are no open TCP ports by default. Only users with OS-level permissions to the socket/pipe can send commands to the daemon.

## 4. Daemon State Machine & The Lifecycle
The daemon maintains its own process-level state machine, independent of the workflows it runs. 
The typical lifecycle is:
1. `Stopped` -> `fenix daemon start`
2. `Starting`
3. `Recovering Journals` (Replaying WAL for Kernel and Runtime)
4. `Loading Plugins` (Handshaking with external gRPC binaries)
5. `Initializing Runtime`
6. `Healthy` (Listening for CLI commands)
7. `Degraded` (e.g., connection lost to Claude Driver)
8. `Stopping`

Because the daemon holds state, `fenix doctor` becomes a powerful tool to inspect the daemon's internal health without running a workflow.

## 5. Streaming Output by Default
Because workflows can take minutes or hours to complete, CLI commands do not block waiting for a synchronous response.
When a user runs `fenix build`, the CLI calls `RPC Build()`. The Daemon creates the Workflow and returns an acknowledgement. The CLI then immediately subscribes to a gRPC Server Stream (`RPC StreamEvents()`), piping real-time EventBus payloads (e.g., `StageStarted`, `InferenceChunkReceived`) straight to standard output with beautiful formatting.

## 6. The Embedded Mode (`--embedded`)
While the Daemon pattern is perfect for local development, it is hostile to CI/CD environments (GitHub Actions, GitLab CI) where spinning up background services for a single, ephemeral job is overkill.
Fénix supports an escape hatch: `fenix build --embedded`.

In embedded mode, the CLI bypasses the gRPC network boundary, directly instantiates the Kernel and Runtime in its own memory space, executes the workflow, and exits cleanly. This is the **only** scenario where the CLI process imports internal logic.

## 7. Command Taxonomy
The CLI verbs are logically grouped by domain, enforcing the idea that the user is interacting with an operating system.

### Workspace Commands
- `fenix init`: Scaffolds `_fenix/` and `workspace.json`.
- `fenix clean`: Wipes `_fenix/cache/` and `_fenix/temp/`.
- `fenix resume`: Restores state from `_fenix/journal/`.
- `fenix doctor`: Queries the Daemon state and Plugin health.

### Workflow Commands
- `fenix run <workflow>`: Dispatches a workflow to the daemon.
- `fenix pause <workflow_id>`: Halts execution.
- `fenix status <workflow_id>`: Inspects the Runtime projection.
- `fenix logs <workflow_id>`: Streams live events.

### Artifact & Plugin Commands
- `fenix artifacts list / show / graph`
- `fenix plugins install / remove / verify`

## 8. Conclusion
**The CLI is never part of the execution engine. It is merely a transport client for the Runtime Control Plane.**
By locking this rule into SPEC-206, Fénix guarantees that any future interface—be it a VSCode Extension, a Web Dashboard, or a REST API—can interact with the exact same Fénix Daemon using the exact same gRPC Protobufs, preserving the single source of truth established in the 100-Series.
