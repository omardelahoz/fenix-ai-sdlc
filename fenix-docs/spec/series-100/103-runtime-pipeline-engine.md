# SPEC-103: Runtime Orchestration Engines (Workflow, Pipeline, Stage)

## 1. Executive Summary
A fatal flaw in naive AI architectures is conflating orchestration with execution. If a "Pipeline" directly invokes an AI "Processor", it is forced to manage state, concurrency, error handling, and inference simultaneously.

Fénix adopts a Kubernetes-grade separation of concerns. The Engineering Runtime orchestrates the SDLC through a strict hierarchy of decoupling:
`Workflow Engine -> Pipeline Engine -> Stage Engine -> Processor Runtime -> Inference Engine`.

Processors do not know about Stages. Stages do not know about Pipelines. The infrastructure manages the DAG (Directed Acyclic Graph), while the Processors strictly transform IRs.

## 2. The Orchestration Hierarchy

```text
    [Events: ProductCreated, PR_Opened]
                   │
    +---------------------------------------------------+
    |             WORKFLOW ENGINE                       |
    | (Decides *which* Pipeline to launch based on      |
    |  Typed Events and Conditions)                     |
    +---------------------------------------------------+
                   │
    +---------------------------------------------------+
    |             PIPELINE ENGINE                       |
    | (Manages the DAG of Stages. Orchestrates the      |
    |  sequential evolution of the Engineering IR)      |
    +---------------------------------------------------+
                   │
    +---------------------------------------------------+
    |              STAGE ENGINE                         |
    | (Manages Fan-Out / Fan-In parallelism. Groups     |
    |  Processors like Frontend, Backend, DB)           |
    +---------------------------------------------------+
                   │
    +---------------------------------------------------+
    |           PROCESSOR RUNTIME                       |
    | (Loads PDL. Injects Planner, Reasoner, Validator) |
    +---------------------------------------------------+
                   │
    +---------------------------------------------------+
    |           INFERENCE ENGINE                        |
    | (Abstracts the AI. Connects to Driver Manager)    |
    +---------------------------------------------------+
```

## 3. Workflow Engine: Event-Driven Triggers
The Workflow Engine listens to the Kernel EventBus. It does not run AI; it evaluates events against WDL triggers.

- **Event:** `ArchitectureUpdated`
- **Condition:** `Architecture.status == approved && Requirements.version == latest`
- **Action:** Triggers the `IncrementalDevelopmentPipeline`.

## 4. Pipeline Engine: The DAG Orchestrator
The Pipeline Engine reads a declarative DAG and traverses it. It never executes code; it delegates to the Stage Engine.

```text
Discovery Stage -> Requirements Stage -> Architecture Stage -> Development Stage
```
Between each Stage, the Pipeline Engine commits a **Deep Pipeline Snapshot** to the Kernel VFS. This snapshot contains the current IR, Journal Pointer, PCB, and Handle state. This enables Fénix to survive total power loss and resume exactly where it left off, akin to Kubernetes state reconciliation.

## 5. Stage Engine: Advanced Concurrency
A Stage contains multiple Processors that can operate concurrently. The Stage Engine manages complex `Fan-Out` and `Fan-In` synchronization.

For example, in the `Development Stage`:
```text
           (Fan-Out)
         / Backend Processor  \
IR ->   -- Frontend Processor --  (Wait-All Join) -> QA Processor
         \ Database Processor /
```
The Stage Engine supports declarative strategies like `wait-all` (all must pass), `wait-first` (race), or `quorum` (majority consensus). It handles assembling the multiple Output IRs back into a single unified IR for the next Stage.

## 6. Processor Runtime: The Generic Executor
The Processor Runtime receives a unified IR from the Stage Engine. It has no domain knowledge. It loads a **PDL (Processor Definition Language)** file which injects the required `Planner`, `Reasoner`, `Validator`, `Transformer`, and `Reporter`.

It explicitly declares its Contracts:
- **Consumes:** `Architecture IR`, `Standards IR`
- **Produces:** `Development IR`, `Code Delta`

By declaring inputs and outputs, the Runtime can statically analyze the Pipeline DAG for correctness before execution begins.

## 7. Inference Engine: The LLM Firewall
The Processor's `Reasoner` never talks to Claude or Gemini directly. It talks to the `Inference Engine`, which parses the `Normalized IR`, queries the `Driver Manager`, and executes the prompt against the optimal LLM. This prevents provider lock-in at the cognitive level.

## 8. Conclusion
By fragmenting the orchestration into Workflow (Events), Pipeline (DAG), Stage (Concurrency), and Processor (Cognitive Transformation), Fénix achieves infinite extensibility. Modifying the SDLC is a matter of editing declarative YAML files, not writing Go code.
