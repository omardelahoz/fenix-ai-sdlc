# SPEC-104: Workflow Definition Language (WDL)

## 1. Executive Summary
The Fénix Operating System is entirely declarative. The core of its orchestration is defined by the **Workflow Definition Language (WDL)**. 
WDL is part of the Fénix Language Family (along with FDL and PDL) and shares the exact same Go Parser and Semantic Analyzer. 

WDL allows architects to define massive, non-linear Directed Acyclic Graphs (DAGs) representing the entire Software Development Life Cycle (SDLC). It introduces advanced concepts such as Typed Events, Fan-Out Concurrency, Human Approval Gates, and Resource Policies.

## 2. The Abstraction Layers
WDL strictly maps to the Fénix Runtime State Model:
- **Workflow:** The entry point. Listens to Kernel events. Contains Pipelines.
- **Pipeline:** A DAG of Stages. Orchestrates the flow of the Engineering IR.
- **Stage:** A concurrency boundary. Groups multiple Processors and orchestrates Fan-Out / Fan-In.

## 3. Workflow Syntax & Typed Events
A workflow triggers based on strict, typed system events rather than arbitrary string matching.

```wdl
workflow StandardSaaSDevelopment

on ProductUpdated
    pipeline MainSaaSPipeline

on PullRequestOpened
    pipeline ReviewPipeline
```

## 4. Pipeline Syntax & The DAG
Pipelines are not linear lists; they are Directed Acyclic Graphs (DAGs). The execution order is determined dynamically by the `depends` keyword.

```wdl
pipeline MainSaaSPipeline

stage Discovery
    processor DiscoveryProcessor

stage Requirements
    depends Discovery
    processor RequirementsProcessor

stage Development
    depends Requirements
    fanout
        strategy wait-all
        processors
            BackendProcessor
            FrontendProcessor
            DatabaseProcessor

stage QA
    depends Development
    fanout
        strategy quorum
        minimum 2
        processors
            SecurityReviewProcessor
            StaticAnalysisProcessor
            ManualReviewProcessor
```

## 5. Fan-Out & Synchronization Strategies
The `Stage Engine` relies on WDL to determine how to synchronize parallel Processors:
- `wait-all`: The Stage does not complete until all Processors return a successful IR.
- `wait-first` (Race): The Stage completes as soon as the fastest Processor returns. The others are automatically cancelled via the Kernel's `CancellationToken`.
- `quorum`: The Stage completes when a minimum number of Processors agree on the Output IR.

## 6. Gates (Human Interaction)
A true SDLC cannot be fully autonomous. WDL introduces `Gates` that pause the Pipeline indefinitely (suspending the `WorkflowInstance`) until human intervention occurs, without consuming Kernel threads.

```wdl
stage Architecture
    depends Requirements
    processor ArchitectureProcessor
    gate HumanApproval
        timeout 48h
        approvers
            SolutionArchitect
            ProductOwner
```
When the `ArchitectureProcessor` finishes, the Pipeline Engine transitions the `StageInstance` to `Suspended`. When the human approves via the CLI or UI, the Kernel emits an `ArchitectureApproved` event, and the Pipeline Engine resumes the DAG.

## 7. Policies & Resources
Processors shouldn't hardcode their retry logic or resource consumption. WDL injects declarative policies that map directly to the Kernel's `ExecutionPackage`.

```wdl
stage Development
    processor BackendProcessor
        policy
            retry
                attempts 3
                backoff exponential
                timeout 30m
            resources
                max-workers 5
                memory 4GB
                priority high
```
The **Scheduler** directly reads these resources when determining when to pull the `ExecutionDescriptor` from the Queue.

## 8. Conclusion
WDL transforms Fénix into an orchestration powerhouse equivalent to Kubernetes or Temporal. By separating the definition of the SDLC (WDL) from the definition of the Processor (PDL), an organization can orchestrate anything from a full product discovery to a simple test regeneration using exactly the same deterministic engine.
