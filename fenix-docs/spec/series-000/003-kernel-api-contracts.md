# SPEC-003: Kernel API Contracts

## 1. Executive Summary
The Fénix Kernel is an **AI-Native Engineering Microkernel**. In a microkernel architecture, components do not hold direct references to each other. They communicate strictly through Inter-Process Communication (IPC) and System Calls (`Syscalls`).

This specification defines the strict internal contracts (Interfaces in Go) for the Fénix Kernel, including the IPC bus, the Virtual File System (VFS), the Semantic Index, the Scheduler (with its Dispatcher), and the AI Device Drivers.

## 2. The Core Architecture (IPC & Syscalls)
The Kernel outlaws direct component coupling. Managers communicate asynchronously via IPC. Executors (Agents) interact with Kernel services exclusively via `Syscalls`.

```go
package kernel

import "context"

// IPC replaces the old EventBus. It allows for future implementations like NATS, gRPC, or local channels.
type IPC interface {
    Request(ctx context.Context, msg Message) (Message, error)
    Notify(ctx context.Context, msg Message) error
    Broadcast(ctx context.Context, msg Message) error
    Reply(ctx context.Context, msgId string, reply Message) error
}

type Message struct {
    Id      string
    Type    MessageType
    Payload interface{}
}

type MessageType string

// KernelAPI defines the System Calls available to Executors.
type KernelAPI interface {
    Call(ctx context.Context, syscall SyscallType, args ...interface{}) (interface{}, error)
}

type SyscallType string

const (
    SysReadArtifact    SyscallType = "ReadArtifact"
    SysWriteArtifact   SyscallType = "WriteArtifact"
    SysQuerySemantic   SyscallType = "QuerySemantic"
    SysCreateTask      SyscallType = "CreateTask"
    SysScheduleTask    SyscallType = "ScheduleTask"
    SysWait            SyscallType = "Wait"
    SysSuspend         SyscallType = "Suspend"
    SysResume          SyscallType = "Resume"
    SysPublish         SyscallType = "Publish"
    SysSubscribe       SyscallType = "Subscribe"
    SysAcquireResource SyscallType = "AcquireResource"
    SysReleaseResource SyscallType = "ReleaseResource"
)
```

## 3. Scheduler & Dispatcher API
The Execution Planner thinks; the Scheduler orders; the Dispatcher decides who works.

```go
package kernel

import "context"

// Scheduler only queues lightweight descriptors. It does not build PCBs.
type Scheduler interface {
    Submit(ctx context.Context, desc ExecutionDescriptor) error
    Cancel(ctx context.Context, descId string) error
}

// Dispatcher assigns queued descriptors to available Workers based on Affinity/NUMA.
type Dispatcher interface {
    // NotifyAvailable is called by the WorkerPool when a thread is free.
    NotifyAvailable(workerId string) 
    AssignNext(ctx context.Context) (*ExecutionDescriptor, string, error)
}

// WorkerPool manages the physical goroutines/threads.
type WorkerPool interface {
    RegisterWorker(workerId string, capabilities []string)
    MarkBusy(workerId string)
    MarkIdle(workerId string)
}
```

## 4. Memory Manager & MMU API
The Memory Manager includes a Memory Management Unit (MMU) responsible for mapping AST nodes into RAM and sharing caches without duplication.

```go
package kernel

import "context"

type MemoryManager interface {
    // BuildSnapshot materializes the PCB only when a Worker is ready.
    BuildSnapshot(ctx context.Context, desc ExecutionDescriptor) (*ExecutionPackage, error)
}

// SnapshotMMU handles the low-level memory operations for L0-L3 caches.
type SnapshotMMU interface {
    Map(nodeId string) (*MemorySegment, error)
    Unmap(nodeId string) error
    Pin(nodeId string) error
    Release(nodeId string) error
    Share(nodeId string, workerIds []string) error
    CopyOnWrite(nodeId string) (*MemorySegment, error)
}
```

## 5. Execution Manager & AI Drivers API
The Execution Manager oversees the lifecycle of the `ExecutionHandle`. It delegates AI inference to generic `AIDriver` interfaces, treating LLMs exactly like hardware device drivers.

```go
package kernel

import "context"

type ExecutionManager interface {
    // Execute takes a finalized PCB and runs it, returning the mutable tracking handle.
    Execute(ctx context.Context, pkg ExecutionPackage) (*ExecutionHandle, error)
}

// AIDriver abstracts Claude, Gemini, OpenAI, Ollama, etc.
type AIDriver interface {
    Init(config DriverConfig) error
    Invoke(ctx context.Context, pcb ExecutionPackage) (DriverResponse, error)
    Stream(ctx context.Context, pcb ExecutionPackage) (<-chan DriverResponseChunk, error)
    HealthCheck(ctx context.Context) error
}
```

## 6. VFS & Semantic Index API
These replace the old "Artifact Store" and "Knowledge Graph", acting as the File System and Semantic Memory Map of the Kernel.

```go
package kernel

import "context"

// VFS (Virtual File System) reads, writes, versions, and indexes data regardless of backend (Git, Local, S3).
type VFS interface {
    Read(ctx context.Context, path string) ([]byte, error)
    Write(ctx context.Context, path string, data []byte) error
    CommitSnapshot(ctx context.Context, handleId string) (string, error) // Returns a commit hash
    Mount(backend VFSBackend) error
}

// SemanticIndex replaces the Graph. It responds to queries and triggers invalidations.
type SemanticIndex interface {
    UpdateIndex(ctx context.Context, commitHash string) error
    QueryDependencies(ctx context.Context, semanticQuery string) ([]string, error)
    MarkDirty(ctx context.Context, nodeIdentifier string) error // Triggers IPC Broadcast for Cache Invalidation
}
```

## 7. Conclusion
By defining Fénix as a true Microkernel, all components become perfectly decoupled. The IPC bus allows distributed execution out-of-the-box, the Syscall interface prevents Agents from accessing internal Kernel state, and the AI Driver model ensures Fénix can swap underlying LLMs instantly without altering core architecture.
