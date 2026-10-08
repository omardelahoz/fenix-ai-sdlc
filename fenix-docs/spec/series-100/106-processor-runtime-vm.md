# SPEC-106: Processor Runtime Virtual Machine (VM)

## 1. Executive Summary
A fatal flaw in most AI agent frameworks is hardcoding domain logic into the agent runner. In Fénix, **a Processor is not code; it is a declarative workload executed by a Virtual Machine**.

The Processor Runtime VM is the equivalent of the JVM (Java) or CLR (.NET). It contains absolutely zero business logic (no knowledge of Backends, Security, or QAs). Its sole responsibility is to parse a Processor Manifest (PMF), inject dependencies via a Capability Registry, and orchestrate the workload through a strict execution lifecycle.

## 2. The ProcessorContext
While the Microkernel strictly operates on a generic `ExecutionPackage`, the Processor Runtime VM operates on the `ProcessorContext`. This context bridges the Kernel's blindness with the Runtime's high-level orchestration state.

```go
type ProcessorContext struct {
    // Orchestration State
    WorkflowInstanceID string
    PipelineInstanceID string
    StageInstanceID    string

    // The Kernel's view of the workload
    Package            kernel.ExecutionPackage

    // Typed Engineering IRs
    InputIR            ir.EngineeringIR
    OutputIR           ir.EngineeringIR

    // Dependency Injection
    Services           RuntimeServices
    Capabilities       CapabilityRegistry
}
```

## 3. Dependency Injection & Capability Registry
Processors never instantiate their own services. The VM handles all Dependency Injection (DI) through two registries:

### A. Runtime Services
Standard OS-level services injected into every context:
- `SemanticQuery`: Allows the processor to query the Knowledge Graph.
- `Journal`: Exposes Kernel VFS checkpointing.
- `Logger` & `Metrics`: Standardized observability.
- `InferenceEngine`: The gateway to LLMs.

### B. Capability Registry
The VM resolves the PMF's declarative `capabilities` to concrete plugin implementations.
- `reasoning` -> Solved by `ClaudePlugin` or `GeminiPlugin`
- `validation` -> Solved by `SMT Solver` or `JSONSchemaValidator`
- `transformation` -> Solved by `GoAstTransformer`

If a PMF requires a capability that is not registered in the VM, execution halts before a single Kernel thread is consumed.

## 4. The Strict Execution Lifecycle
The VM enforces a deterministic state machine for every Processor workload. The loaded plugins execute within these Lifecycle Hooks:

1. **Initialize:** Parse the PMF and establish the `ProcessorContext`.
2. **ResolveDependencies:** Verify all required `Capabilities` are present in the Registry.
3. **AcquireCapabilities:** Instantiate the resolved plugins.
4. **PreValidate:** Check `Contracts`. Does the `InputIR` match the PMF's `consumes` list?
5. **Reason:** Invoke the `InferenceEngine` via the Reasoning plugin.
6. **Transform:** Execute transformations on the resulting IR (e.g., Code generation).
7. **Validate:** The Validation plugin ensures the `OutputIR` adheres strictly to the Fénix Type System.
8. **Report:** Format trace logs and telemetry.
9. **ReleaseResources & Dispose:** Free memory and gracefully terminate the VM thread.

## 5. Observability & Telemetry (Out-of-the-Box)
Because the VM controls the entire lifecycle, it emits standard telemetry events at every state transition. 
Events like `ProcessorStarted`, `CapabilityResolved`, `InferenceCompleted`, and `ValidationFailed` are pushed to the Kernel EventBus automatically. 

This gives Fénix Temporal-grade observability without requiring the Processor author to write a single logging statement.

## 6. Conclusion
The Processor Runtime VM is the engine of Fénix intelligence. By abstracting the execution of AI workloads into a strict VM lifecycle, Fénix ensures that the underlying Microkernel remains frozen, while the Runtime remains infinitely extensible through declarative manifests and plugins.
