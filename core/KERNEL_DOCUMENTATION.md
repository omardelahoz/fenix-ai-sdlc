# 📦 Core Module - Detailed Documentation

## Overview

The core module is the consolidated foundation of Fénix, containing the Kernel, Runtime, Inference Engine, and Parser. This module follows the architectural guidelines defined in [SPEC-200](../fenix-docs/spec/series-200/200-repository-layout.md).

## Architecture

```
core/
├── internal/          # Private application code (not importable by plugins)
│   └── kernel/        # The execution engine (SPEC-000)
│       ├── kernel.go
│       ├── execution_manager.go
│       ├── memory_manager.go
│       ├── state_machine.go
│       ├── scheduler.go
│       ├── claude_adapter.go
│       ├── antigravity_adapter.go
│       ├── ollama_adapter.go
│       ├── provider_factory.go
│       └── provider_example.go
└── pkg/               # Stable internal library code
    └── contracts/     # Shared Go structs and interfaces
        ├── execution.go
        ├── execution_handle.go
        ├── kernel.go
        └── provider.go
```

## Packages

### pkg/contracts

Contains shared data structures and interfaces used across the Kernel, Runtime, and other subsystems.

#### execution.go

Defines the immutable Process Control Block (PCB) of the Fénix Kernel:

- **ExecutionPackage** - Immutable contract for task execution
- **PackageMetadata** - Versioning and tracing information
- **TaskDescriptor** - Task description
- **ExecutorDescriptor** - Abstract representation of agents/skills/plugins
- **ExecutionSnapshot** - Frozen state of the repository
- **ExecutionPolicies** - Execution constraints and retry logic
- **ResourceAllocation** - Budget and deadlines
- **Contracts** - Expected output format and validation

**Key Feature:** ExecutionPackage is completely immutable. Any state change creates a new version via `NewVersion()`.

#### execution_handle.go

Defines the mutable runtime state tracking:

- **ExecutionHandle** - Tracks live execution state
- Manages: Attempt count, StartedAt/EndedAt, Duration, CurrentPhase, Status
- Supports: Cancel, MarkCompleted, MarkFaulted, IncrementAttempt

**Key Feature:** Separates immutable ExecutionPackage from mutable runtime state.

#### kernel.go

Defines the core Kernel interfaces:

- **Kernel** - Main interface for the Fénix Kernel
- **MemoryManager** - Manages memory references and L0 cache
- **ExecutionManager** - Manages ExecutionHandle lifecycle
- **StateMachine** - Manages Kernel state transitions
- **ContextModel** - Execution context for tasks

#### provider.go

Defines AI provider abstraction:

- **AIProvider** - Interface for all AI providers (Claude, Antigravity, etc.)
- **Prompt** - Prompt structure (system message, user message, messages)
- **Response** - Response structure (text, finish reason, metadata)
- **ProviderConfig** - Generic provider configuration
- **ClaudeConfig** - Claude-specific configuration
- **AntigravityConfig** - Antigravity-specific configuration
- **OllamaConfig** - Ollama-specific configuration

### internal/kernel

Implements the Fénix Kernel (SPEC-000).

#### kernel.go

Main Kernel implementation:

- **FenixKernel** - Implements Kernel interface
- **Initialize()** - Initializes subsystems and transitions to Running state
- **Shutdown()** - Gracefully shuts down all executions
- **Execute()** - Executes an ExecutionPackage and returns a handle
- **Cancel()** - Cancels an execution by package ID
- **GetHandle()** - Retrieves an ExecutionHandle
- **ListHandles()** - Lists all active handles
- **State()** - Returns current Kernel state
- **OnTransition()** - Registers callbacks for state transitions

**State Transitions:**
```
Uninitialized → Initializing → Running → Suspended → ShuttingDown → Shutdown
```

#### execution_manager.go

Manages ExecutionHandle lifecycle:

- **ExecutionManagerImpl** - Implements ExecutionManager interface
- **CreateHandle()** - Creates a new ExecutionHandle for a package
- **UpdateHandle()** - Updates an existing ExecutionHandle
- **GetHandle()** - Retrieves a handle by ID
- **ListHandles()** - Lists all active handles
- **Cleanup()** - Removes completed/faulted handles

**Key Feature:** Prevents duplicate handles for the same package ID.

#### memory_manager.go

Manages memory references and L0 cache:

- **MemoryManagerImpl** - Implements MemoryManager interface
- **Resolve()** - Resolves memory references with L0 cache
- **Cache()** - Caches an ExecutionSnapshot
- **CacheMetadata()** - Caches package metadata separately
- **GetMetadata()** - Retrieves cached metadata
- **Invalidate()** - Invalidates cached snapshots and related data
- **InvalidateASTNode()** - Invalidates a specific AST node
- **Cleanup()** - Removes expired snapshots
- **Stats()** - Returns cache statistics (hits, misses, hit rate)

**L0 Cache Features:**
- AST node caching for fast resolution
- Metadata caching for quick access
- Automatic cleanup based on TTL
- Hit/miss tracking and statistics

#### state_machine.go

Manages Kernel state transitions:

- **StateMachineImpl** - Implements StateMachine interface
- **Current()** - Returns current state
- **Transition()** - Attempts to transition to a new state
- **CanTransition()** - Checks if a transition is valid
- **OnTransition()** - Registers callbacks for state transitions

**Valid Transitions:**
- Uninitialized → Initializing
- Initializing → Running, Shutdown
- Running → Suspended, ShuttingDown
- Suspended → Running, ShuttingDown
- ShuttingDown → Shutdown
- Shutdown → (terminal)

#### scheduler.go

Schedules and manages task execution:

- **Scheduler** - Schedules ExecutionPackages
- **Schedule()** - Adds a task to the queue
- **ScheduleDAG()** - Schedules multiple packages as a DAG
- **Run()** - Starts the scheduler loop
- **Cancel()** - Cancels a scheduled or running task
- **GetQueueSize()** - Returns current queue size
- **GetRunningCount()** - Returns number of running tasks
- **Stats()** - Returns scheduler statistics

**Features:**
- Queue-based task scheduling
- Concurrent execution support
- Priority mapping
- Background worker pool

#### claude_adapter.go

Implements AIProvider for Anthropic Claude:

- **ClaudeAdapter** - Full Claude API integration
- **Chat()** - Sends prompts to Claude API
- **Stream()** - Streams responses (fully implemented with SSE events)
- **ProviderName()** - Returns "claude"
- **AvailableModels()** - Lists available Claude models

**Streaming Features:**
- Real-time streaming via Claude's SSE API
- Handles multiple event types (message_start, content_block_delta, message_stop)
- Error handling in real-time
- Automatic completion detection

**Supported Models:**
- claude-3-5-sonnet-20241022
- claude-3-5-haiku-20241022
- claude-3-opus-20240229
- claude-3-sonnet-20240229
- claude-3-haiku-20240307

#### antigravity_adapter.go

Implements AIProvider for Antigravity platform:

- **AntigravityAdapter** - Antigravity API integration
- **Chat()** - Sends prompts to Antigravity API
- **Stream()** - Streams responses (fully implemented with SSE)
- **ProviderName()** - Returns "antigravity"
- **AvailableModels()** - Lists available Antigravity models

**Streaming Features:**
- Real-time streaming via SSE (Server-Sent Events)
- Parses SSE format (data: {...})
- Handles [DONE] marker for completion
- Error handling in real-time

**Features:**
- Workspace and project organization
- Specialized models for software engineering

#### ollama_adapter.go

Implements AIProvider for Ollama (local LLMs):

- **OllamaAdapter** - Ollama API integration
- **Chat()** - Sends prompts to local Ollama instance
- **Stream()** - Streams responses (fully implemented)
- **ProviderName()** - Returns "ollama"
- **AvailableModels()** - Queries `/api/tags` for available models

**Benefits:**
- Complete privacy (data never leaves machine)
- No API costs
- Supports various models (Llama, Mistral, CodeLlama, etc.)
- Offline operation

#### provider_factory.go

Factory for creating AIProvider instances:

- **ProviderFactory** - Creates providers dynamically
- **CreateProvider()** - Creates provider by name and config
- **CreateProviderFromExecutionPackage()** - Creates provider from ExecutionPackage
- **ListAvailableProviders()** - Lists all available providers

**Supported Providers:**
- claude
- antigravity
- ollama
- openai (pending)
- gemini (pending)

#### provider_example.go

Examples of provider usage:

- **ExampleUsage()** - Demonstrates using Claude, Antigravity, and Ollama
- **ExampleDynamicProviderSelection()** - Demonstrates dynamic provider selection
- Shows different strategies: by task type, fallback chain, cost optimization

#### streaming_example.go

Examples of streaming usage:

- **ExampleStreaming()** - Demonstrates streaming with all providers
- **ExampleStreamingWithCancellation()** - Shows how to cancel streaming requests
- **ExampleStreamingWithBuffer()** - Demonstrates collecting stream into buffer
- **ExampleStreamingWordByWord()** - Word-by-word streaming example
- **ExampleStreamingComparison()** - Compares streaming vs non-streaming latency

## Usage Examples

### Basic Kernel Usage

```go
package main

import (
    "context"
    "time"
    
    "github.com/omardelahoz/fenix-ai-sdlc/core/internal/kernel"
    "github.com/omardelahoz/fenix-ai-sdlc/core/pkg/contracts"
)

func main() {
    // Create kernel
    k := kernel.NewKernel()
    
    // Initialize
    ctx := context.Background()
    if err := k.Initialize(ctx); err != nil {
        panic(err)
    }
    defer k.Shutdown(ctx)
    
    // Create execution package
    pkg := &contracts.ExecutionPackage{
        Metadata: contracts.PackageMetadata{
            ID:      "pkg-001",
            Version: "v1",
        },
        Task: contracts.TaskDescriptor{
            ID:   "task-001",
            Type: "GenerateCode",
        },
        Executor: contracts.ExecutorDescriptor{
            ID:   "executor-001",
            Type: "Agent",
        },
        Resources: contracts.ResourceAllocation{
            Provider: "claude-3.5-sonnet",
            Deadline: time.Now().Add(time.Hour),
        },
    }
    
    // Execute
    handle, err := k.Execute(ctx, pkg)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Execution completed: %+v\n", handle)
}
```

### Using AI Providers

```go
// Create provider factory
factory := kernel.NewProviderFactory()

// Configure Claude
claudeConfig := &contracts.ClaudeConfig{
    ProviderConfig: contracts.ProviderConfig{
        APIKey:      "sk-ant-xxx",
        Model:       "claude-3-5-sonnet-20241022",
        MaxTokens:   100000,
        Temperature: 0.7,
    },
    Region: "us-east-1",
}

// Create provider
provider, err := factory.CreateProvider("claude", claudeConfig)
if err != nil {
    panic(err)
}

// Send prompt
prompt := contracts.Prompt{
    SystemMessage: "You are a software engineer.",
    UserMessage:   "Generate a REST API.",
}

response, err := provider.Chat(context.Background(), prompt)
if err != nil {
    panic(err)
}

fmt.Println(response.Text)
```

### Using Scheduler

```go
// Create mock executor
executor := &mockExecutor{
    executeFunc: func(ctx context.Context, pkg *contracts.ExecutionPackage) (*contracts.ExecutionHandle, error) {
        return contracts.NewExecutionHandle(pkg), nil
    },
}

// Create scheduler
scheduler := kernel.NewScheduler(executor)

// Schedule tasks
for i := 0; i < 10; i++ {
    pkg := &contracts.ExecutionPackage{
        Metadata: contracts.PackageMetadata{
            ID: "pkg-" + string(rune('0'+i)),
        },
        // ... other fields
    }
    
    if err := scheduler.Schedule(ctx, pkg); err != nil {
        panic(err)
    }
}

// Start scheduler
go scheduler.Run(ctx)
```

## Testing

### Running Tests

```bash
# Run all tests
cd core
go test ./...

# Run specific package tests
go test ./pkg/contracts/...
go test ./internal/kernel/...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...
```

### Test Coverage

- **execution_test.go** - Tests for ExecutionPackage and ExecutionHandle
- **kernel_test.go** - Integration tests for Kernel
- **execution_manager_test.go** - Tests for ExecutionManager
- **memory_manager_test.go** - Tests for MemoryManager (pending)
- **state_machine_test.go** - Tests for StateMachine (pending)
- **scheduler_test.go** - Tests for Scheduler (pending)
- **concurrency_test.go** - Concurrent execution tests

## Invariants

The Kernel enforces the following invariants:

### ExecutionPackage Invariants
- SP-001: ExecutionPackage is completely immutable
- SP-002: NewVersion() creates a new version, never mutates in-place
- SP-003: Validation must pass before execution

### Kernel State Invariants
- SCH-001: Only one state transition at a time
- SCH-002: Cannot transition from ShuttingDown back to Running
- SCH-003: Cannot execute in Shutdown state

### Memory Manager Invariants
- MM-001: Cache lookups are thread-safe
- MM-002: Invalidation cascades to related data
- MM-003: Cleanup respects TTL

## Performance Considerations

### Memory Management
- L0 cache for AST nodes reduces memory allocations
- Copy-on-write for snapshots reduces memory footprint
- Automatic cleanup prevents memory leaks

### Concurrency
- Lock-free reads for snapshots
- Mutex-protected writes for mutations
- Worker pool for parallel execution

### Scheduling
- Queue-based scheduling prevents overwhelm
- Priority mapping ensures critical tasks run first
- Background workers for long-running tasks

## Integration Points

### With Runtime (SPEC-100)
- Kernel provides ExecutionPackages to Runtime
- Runtime returns ExecutionPackages for validation
- Memory Manager serves Runtime's semantic needs

### With Language System (SPEC-300)
- Kernel caches AST nodes from Language System
- Language System provides semantic validation
- Shared contracts define the interface

### With Compiler Host (SPEC-400)
- Kernel executes within Host's Worker Pool
- Kernel respects Host's Resource Manager
- Kernel reports telemetry to Host

## Future Enhancements

### Short Term
- [ ] Implement streaming for Claude and Antigravity adapters
- [ ] Add OpenAI adapter
- [ ] Add Gemini adapter
- [ ] Implement disk-based cache persistence
- [ ] Add more sophisticated cache eviction policies

### Long Term
- [ ] Implement distributed memory management
- [ ] Add snapshot compression
- [ ] Implement cross-node scheduling
- [ ] Add provider-specific optimization hints
- [ ] Implement automatic provider selection based on task characteristics

## References

- [SPEC-000](../fenix-docs/spec/series-000/) - Kernel API Contracts
- [SPEC-001](../fenix-docs/spec/series-000/001-execution-package.md) - Execution Package Specification
- [SPEC-003](../fenix-docs/spec/series-000/003-kernel-api-contracts.md) - Kernel API Contracts
- [SPEC-004](../fenix-docs/spec/series-000/004-kernel-state-machine.md) - Kernel State Machine
- [INTEGRATION.md](../fenix-docs/INTEGRATION.md) - Provider Integration Guide
