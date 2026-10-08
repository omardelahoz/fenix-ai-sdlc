# SPEC-200: Repository Layout

## 1. Executive Summary
As Fénix transitions from architectural design to code implementation, the physical structure of the repository becomes the first line of defense against technical debt and circular dependencies. 
Fénix adopts a **Consolidated Monorepo Layout** managed by `go.work`. To minimize Go module overhead, all core execution layers (Kernel, Runtime, Inference, Parser) are grouped into a single `/core` module, while retaining strict encapsulation via standard Go package visibility rules (`internal/` vs `pkg/`).

## 2. The Consolidated Monorepo Philosophy
Maintaining a separate `go.mod` for `kernel`, `runtime`, and `parser` creates a massive operational burden with no benefit, as these components are inherently versioned and released together. 
Fénix consolidates the foundation into a single `core` module. Independent `go.mod` files are reserved strictly for components that must evolve independently: the CLI (`cmd/fenix`), Plugins, and public SDKs.

## 3. Root Directory Structure
The repository strictly adheres to the following structure:

```text
/fenix
├── go.work                 # Go workspace definition linking all modules
├── Taskfile.yml            # Primary build and automation runner
│
├── /cmd                    # Main applications for this project
│   └── /fenix              # [Module] The core CLI binary (e.g., `fenix build`)
│
├── /core                   # [Module] The consolidated Fénix engine
│   ├── /internal           # Private application code (Not importable by plugins)
│   │   ├── /kernel         # The execution engine
│   │   ├── /runtime        # The orchestration engine
│   │   ├── /inference      # The Inference Engine compilation pipeline
│   │   └── /parser         # FDL, WDL, PMF lexers and parsers
│   │
│   └── /pkg                # Stable internal library code
│       ├── /contracts      # Shared Go structs (ExecutionPackage, ProcessorContext)
│       ├── /api            # Generated gRPC/Protobuf code
│       └── /schemas        # JSON Schemas for validation
│
├── /proto                  # Raw gRPC .proto definitions (The ABI contract)
│
├── /plugins                # Out-of-process executable plugins
│   ├── /claude             # [Module] Anthropic Claude Driver
│   └── /openai             # [Module] OpenAI Driver
│
├── /sdk                    # [Module] Developer SDKs for creating plugins
│
├── /tools                  # Dev utilities (schema-generator, mock-generator)
│
├── /docs                   # Documentation, ADRs, and the SPEC series
├── /examples               # Example FDL projects and workflows
└── /test                   # E2E integration testing suite
```

## 4. The Enforced Boundaries (The Contracts Package)
The most critical architectural decision within `/core` is the separation of implementation (`/internal`) and shared contracts (`/pkg/contracts`).

**The Problem:** The `Runtime` needs to send an `ExecutionPackage` to the `Kernel`. If the `ExecutionPackage` struct is defined in `/internal/kernel`, the `Runtime` must import `kernel`. If the `Kernel` needs to read a `ProcessorContext` from the `Runtime`, it must import `runtime`. This creates a forbidden cyclic dependency.

**The Solution:** Neither `kernel` nor `runtime` import each other. 
All shared structures, internal APIs, and interface definitions live in `/core/pkg/contracts`. 
- `/core/internal/kernel` imports `/core/pkg/contracts`.
- `/core/internal/runtime` imports `/core/pkg/contracts`.

This guarantees that both massive subsystems can be compiled, tested, and refactored in complete isolation.

## 5. API vs Contracts
Inside `/core/pkg`, there is a strict division:
- `/pkg/contracts` holds pure, idiomatic Go structs (e.g., `Artifact`, `ExecutionHandle`).
- `/pkg/api` holds the auto-generated code from the `/proto` directory.
This ensures that internal Go domain models evolve independently from the strict gRPC wire protocol used to communicate with Out-of-Process plugins.

## 6. Plugin Segregation
The `/plugins` directory does not compile into the main `fenix` binary in `/cmd/fenix`.
Because Fénix uses an Out-of-Process gRPC architecture (SPEC-109), each folder inside `/plugins` is its own independent `go.mod` and yields its own binary (e.g., `fenix-plugin-claude.exe`). The core Fénix daemon launches these binaries as child processes at runtime.

## 7. Conclusion
By consolidating the core engine into a single module while maintaining rigorous package-level encapsulation, Fénix achieves a physically robust codebase. The repository structure mirrors the conceptual architecture: Definitions (`/docs`), Runtime/Kernel (`/core`), Plugins (`/plugins`), providing unparalleled clarity for future contributors.
