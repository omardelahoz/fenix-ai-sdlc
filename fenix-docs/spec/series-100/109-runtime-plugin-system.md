# SPEC-109: Runtime Plugin System (The Extensibility Boundary)

## 1. Executive Summary
If a framework allows third-party extensions to crash the core daemon, it is not an Operating System. Fénix adopts a strict **Out-of-Process, Manifest-Driven Plugin Architecture**. 
To maintain the stability of the Frozen Microkernel, the Runtime enforces a hard boundary: **No arbitrary code is loaded into the Fénix process.** 
Plugins are managed by a `Plugin Host` and communicate exclusively via a stable gRPC interface, allowing extensions to be written in Go, Rust, C++, or Python without threatening system integrity.

## 2. The Final Architecture Stack
This SPEC solidifies the macroscopic architecture of Fénix:

```text
+-----------------------------------------------------------+
| 1. Definition Layer (FDL, WDL, PMF, Schemas)              |
+-----------------------------------------------------------+
                             │
+-----------------------------------------------------------+
| 2. Runtime Layer (Workflow, Pipeline, VM, Inference)      |
+-----------------------------------------------------------+
                             │
+-----------------------------------------------------------+
| 3. Plugin Layer (Registry, Host, gRPC Isolation)          |
+-----------------------------------------------------------+
                             │
+-----------------------------------------------------------+
| 4. Microkernel (Syscalls, IPC, VFS, Memory, Scheduler)    |
+-----------------------------------------------------------+
```
The **Plugin Layer** is the *only* extensibility boundary. To add new AI drivers, SDLC methodologies, or FDL language schemas, one writes a plugin, never a Kernel patch.

## 3. Out-of-Process Isolation (gRPC)
Fénix explicitly bans Go native plugins, DLLs, and reflection-based dynamic loading due to their fragility (ABI incompatibilities, memory leaks, panics).

Executable Plugins (like an `SMT Solver` for validation) are spawned as isolated child processes by the **Plugin Host**.
```text
[Processor Runtime VM] -> [gRPC Client] ==== (IPC) ==== [gRPC Server] -> [Rust/Python/Go Plugin]
```
- **Resilience:** If a plugin encounters a Segfault, only its process dies. The `Plugin Host` catches the disconnect, logs the error, and can optionally restart the plugin without crashing Fénix.
- **Resource Limits:** Because plugins are processes, the OS can enforce strict memory (cgroups) and CPU bounds on them.

## 4. The Universal Plugin Manifest
Every extension in Fénix (be it a Processor, Driver, or Capability) shares a universal base manifest schema to guarantee traceability and security.

```yaml
plugin:
  id: "com.fenix.capabilities.claude"
  version: "1.2.0"
  type: "Capability"
  author: "Fénix Core Team"
  description: "Anthropic Claude reasoning engine."

dependencies:
  api: "1.0"
  runtime: ">= 1.2"
  kernel: ">= 1.0"

security:
  checksum: "sha256:8f43434..."
  signature: "MEYCIQ..." # Enterprise PKI verification
```

## 5. Capability-First Registry
The Runtime achieves total decoupling by querying a **Capability Registry**, not a plugin name.
A Processor Manifest (PMF) declares: `capabilities: [reasoning]`.
The Registry evaluates installed plugins:
1. `ClaudePlugin` (provides `reasoning`, version `1.2.0`)
2. `GeminiPlugin` (provides `reasoning`, version `1.0.0`)

The Registry resolves the best candidate (based on configuration or priority) and returns a gRPC connection to the Runtime. The Runtime is completely agnostic to whether it is talking to Claude or Gemini.

## 6. Pluggable Language Schemas
FDL (Fénix Definition Language) is not a closed ecosystem. The Plugin System allows for **Schema Plugins**.
An industry consortium could release `Automotive.fdl` (containing safety-critical schemas like ISO26262 requirements). Dropping this plugin into Fénix instantly allows the unified Parser to recognize and validate the new nodes, making Fénix infinitely adaptable to niche engineering domains.

## 7. Strict Plugin Lifecycle
Plugins do not just "load". The Plugin Host manages a deterministic state machine for every extension:
`Discover -> Validate Signature -> Resolve Dependencies -> Load (Spawn Process) -> Initialize -> Healthy -> Serving -> Stopping -> Unloaded`.
If a plugin fails dependency resolution (e.g., it requires `kernel >= 2.0` but the system is running `1.0`), it is gracefully rejected before any code executes.

## 8. Conclusion
The Runtime Plugin System completes the Fénix architecture. By isolating extensibility behind a strict gRPC boundary and a capability-first registry, Fénix combines the absolute determinism of a frozen Microkernel with the limitless flexibility of an open ecosystem.
