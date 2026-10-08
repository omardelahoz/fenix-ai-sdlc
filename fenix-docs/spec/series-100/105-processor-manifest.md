# SPEC-105: Processor Manifest (PMF) & The Unified Language Family

## 1. Executive Summary
A massive architectural simplification in Fénix is the realization that **FDL, WDL, and PMF are not distinct languages.** They are specialized schemas (Profiles) of the single **Fénix Definition Language**. 
They share the exact same Lexer, Parser, AST, and Semantic Analyzer in Go. The only difference is the Root Node of the document, akin to how HTML and SVG are both merely XML schemas.

Within this ecosystem, the **Processor Manifest (PMF)** represents the contract of a Processor. Processors are not hardcoded Go classes bound to Claude or Gemini; they are dynamic entities. A PMF declares the Processor's identity, input/output contracts, required capabilities, and inference profiles. The Runtime reads this manifest and injects the appropriate engines to fulfill it.

## 2. The Processor Manifest Abstraction

```text
    +---------------------------------------------------+
    |           PROCESSOR MANIFEST (PMF)                |
    | (Declares Capabilities, Contracts, Profiles)      |
    +---------------------------------------------------+
                              │ (Parsed by Runtime)
                              ▼
    +---------------------------------------------------+
    |              PROCESSOR RUNTIME                    |
    | (Matches PMF Capabilities to installed Plugins)   |
    +---------------------------------------------------+
                              │ (Injects Implementations)
      ┌───────────────┬───────┴───────┬───────────────┐
      ▼               ▼               ▼               ▼
 [Planner Plugin] [Reasoner Plugin] [Validator Plugin] [Transformer]
 (e.g. Go Code)   (e.g. Claude)     (e.g. SMT Solver)  (e.g. Rust)
```

## 3. PMF Syntax & Capabilities
A Processor does not declare *how* it does its job. It declares *what* it is capable of doing. 
If a PMF requires `reasoning`, the Runtime can dynamically use Claude. If Claude fails, it can gracefully fallback to Gemini, because the PMF only asked for a capability, not a specific provider.

### Example PMF
```pmf
processor RequirementsProcessor

version 2.1
api 1.0

dependencies
    runtime >= 1.0
    kernel >= 1.0

capabilities
    - reasoning
    - planning
    - validation
    - transformation
    - synthesis

contracts
    consumes
        - ProductIR
        - StandardsIR
    produces
        - RequirementsIR
    guarantees
        - DeterministicOutput
        - ValidIR
        - Traceability

inference
    profile high_reasoning
        temperature 0.0
        context large
        stream true
```

## 4. Input / Output Contracts (Static Validation)
The `contracts` block (`consumes` and `produces`) is revolutionary for Fénix. 

Before a `WorkflowInstance` is even started, the `Pipeline Engine` can trace the DAG of the WDL and match it against the PMF of every Processor. If a `BackendProcessor` consumes an `ArchitectureIR`, but the preceding stage does not produce an `ArchitectureIR`, the Runtime will throw a **Static Analysis Error** immediately. 

This brings compile-time safety to AI SDLC Orchestration.

## 5. Inference Profiles
The `inference` block ensures that the AI's hyper-parameters are decoupled from the core logic. 

A `CodeGenerationProcessor` might request a `high_creativity` profile, while a `CodeReviewProcessor` requires a `high_reasoning` profile with `temperature 0`. The **Inference Engine** (SPEC-107) translates these semantic profiles into the precise API calls required by the chosen underlying driver (OpenAI, Anthropic, Google).

## 6. Conclusion
By treating Processors as Manifests rather than Code, Fénix achieves true polymorphism. The community can build a **Processor Marketplace** where downloading a new SDLC methodology (e.g., a "Security Audit Processor") is as simple as dropping a `.pmf` file into the workspace. The Runtime handles the rest.
