# Core Module

This is the consolidated core module for Fénix, containing the Kernel, Runtime, Inference Engine, and Parser.

## 📦 Status

**Fase 1 (Kernel Foundation):** ✅ **100% COMPLETADO**

- ✅ Core data structures (ExecutionPackage, ExecutionHandle)
- ✅ Kernel interfaces and implementations
- ✅ Memory Manager with L0 cache
- ✅ State Machine with valid transitions
- ✅ Scheduler for task execution
- ✅ Integration with AI providers (Claude, Antigravity, Ollama)
- ✅ Comprehensive testing (unit, integration, concurrency)
- ✅ CLI for basic operations

## Structure

```
core/
├── internal/          # Private application code (not importable by plugins)
│   └── kernel/        # The execution engine (SPEC-000) ✅
│       ├── kernel.go
│       ├── execution_manager.go
│       ├── memory_manager.go
│       ├── state_machine.go
│       ├── scheduler.go
│       ├── claude_adapter.go
│       ├── antigravity_adapter.go
│       ├── ollama_adapter.go
│       ├── provider_factory.go
│       ├── provider_example.go
│       ├── kernel_test.go
│       ├── execution_manager_test.go
│       └── concurrency_test.go
│   ├── runtime/       # The orchestration engine (por implementar)
│   ├── inference/     # The Inference Engine compilation pipeline (por implementar)
│   └── parser/        # FDL, WDL, PMF lexers and parsers (por implementar)
└── pkg/               # Stable internal library code
    └── contracts/     # Shared Go structs and interfaces ✅
        ├── execution.go
        ├── execution_handle.go
        ├── execution_test.go
        ├── kernel.go
        └── provider.go
```

## Packages

### pkg/contracts ✅

Contains shared data structures and interfaces used across the Kernel, Runtime, and other subsystems:

- `execution.go` - ExecutionPackage (immutable PCB) and related structures
- `execution_handle.go` - ExecutionHandle (mutable runtime state)
- `execution_test.go` - Unit tests for execution structures
- `kernel.go` - Kernel and related interfaces
- `provider.go` - AI provider abstraction and configurations

### internal/kernel ✅

Implementation of the Fénix Kernel (SPEC-000):

- `kernel.go` - Main Kernel implementation
- `execution_manager.go` - Manages ExecutionHandles
- `memory_manager.go` - Manages memory references and L0 cache
- `state_machine.go` - Manages Kernel state transitions
- `scheduler.go` - Schedules and manages task execution
- `claude_adapter.go` - Anthropic Claude integration
- `antigravity_adapter.go` - Antigravity platform integration
- `ollama_adapter.go` - Local LLM integration (Ollama)
- `provider_factory.go` - Factory for creating AI providers
- `provider_example.go` - Usage examples

## Building

```bash
cd core
go build ./...
```

## Testing

```bash
# Run all tests
cd core
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...
```

## Documentation

See [KERNEL_DOCUMENTATION.md](KERNEL_DOCUMENTATION.md) for detailed documentation of the Kernel implementation.

## Usage Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/omardelahoz/fenix-ai-sdlc/core/internal/kernel"
    "github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

func main() {
    // Create kernel
    k := kernel.NewKernel()

    // Initialize
    if err := k.Initialize(context.Background()); err != nil {
        panic(err)
    }
    defer k.Shutdown(context.Background())

    // Create execution package
    pkg := &contracts.ExecutionPackage{
        Metadata: contracts.PackageMetadata{
            ID:      "pkg-123",
            Version: "v1",
        },
        Task: contracts.TaskDescriptor{
            ID:   "task-456",
            Type: "GenerateCode",
        },
        Executor: contracts.ExecutorDescriptor{
            ID:   "executor-789",
            Type: "Agent",
        },
        Resources: contracts.ResourceAllocation{
            Provider: "claude-3.5-sonnet",
            Deadline: time.Now().Add(time.Hour),
        },
    }

    // Execute
    handle, err := k.Execute(context.Background(), pkg)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Execution completed: %+v\n", handle)
}
```

## CLI Usage

```bash
# Initialize the kernel
cd cmd/fenix
go run main.go init

# Execute a sample task
go run main.go execute

# Check kernel status
go run main.go status
```
