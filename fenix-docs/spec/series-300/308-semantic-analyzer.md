# SPEC-308: Semantic Analyzer

## 1. Executive Summary
The **Semantic Analyzer** is the final validation gate in the compilation pipeline before code generation (IR). While the Binder establishes identity and resolves references (SPEC-307), the Semantic Analyzer enforces the rules of the language and the domain.

To maintain a mathematically deterministic compiler while enabling massive enterprise extensibility, the Fénix Semantic Analyzer is split into two distinct architectural layers:
1. **The Core Semantic Analyzer** (Mandatory, internal, defines the language).
2. **The Analyzer Framework** (Pluggable, external, enforces domain rules).

## 2. The Core Semantic Analyzer
The Core is **not** a plugin. It is an immutable, hardcoded part of the compiler responsible for the *Core Language Semantics*. If these rules were pluggable, Fénix would lose determinism and could splinter into incompatible dialects.

To maximize maintainability, the Core operates as a sequence of specialized **Analysis Passes**:
1. Type Inference
2. Conversions & Casting
3. Overload Resolution
4. Accessibility & Scoping
5. Constant Folding Validation

*(Note: The Core may internally delegate control-flow and data-flow analysis—such as definite assignment, uninitialized variables, or nullability flow—to specialized passes. These remain part of the compiler core and are not exposed as pluggable analyzers).*

The Core runs immediately after the Binder. Upon completion, it **Freezes** the `SemanticModel`. From this point forward, the model is strictly read-only for the pluggable ecosystem.

## 3. The Analyzer Framework (Pluggable Rules)
Once the Core finishes and freezes the model, the compiler invokes the **Analyzer Framework**. This is a pluggable pipeline where additional rule sets inspect the `SemanticModel` to enforce style, security, domain compliance, and enterprise policies without touching the compiler source code.

### 3.1 The Analyzer Contract
An Analyzer is a lightweight module governed by strict capability flags, dependencies, and metadata:

```go
type AnalyzerMetadata struct {
    Version  string
    Author   string
    Category AnalyzerCategory
}

type Analyzer interface {
    ID() string
    Metadata() AnalyzerMetadata
    DefaultSeverity() DiagnosticSeverity
    Dependencies() []string // Allows execution ordering (e.g., Security depends on Type Analyzer)
    
    // Capabilities
    SupportsParallel() bool
    SupportsIncremental() bool
    
    Analyze(ctx *AnalysisContext) AnalysisResult
}

type AnalysisContext struct {
    Context       context.Context // Enables fast cancellation in the IDE
    Scope         AnalysisScope   // Current File, Project, or Workspace
    Compilation   *Compilation
    SemanticModel *SemanticModel
    Configuration AnalyzerConfig  // Dynamic rules (e.g., MaxEntityFields=50)
    Diagnostics   *DiagnosticCollector
}

type AnalysisResult struct {
    Metrics    AnalysisMetrics
    Statistics AnalysisStats
}
```

### 3.2 The Semantic-First Read-Only Invariant
To preserve compilation integrity, an Analyzer is strictly **Read-Only**.

**Semantic-First Rule:** Analyzer plugins should be semantic-first. They must consume the `SemanticModel`, `Symbols`, and `Bindings` as their primary API. Access to the Canonical AST, when available, is read-only and intended only for scenarios that require syntactic context (e.g., code style, formatting, or source-aware diagnostics). Semantic correctness must never depend on traversing the AST directly.

An Analyzer **CANNOT**:
- Mutate the AST (e.g., `ctx.AST.Replace(...)`)
- Add or modify symbols (e.g., `ctx.SymbolTable.Add(...)`)
- Mutate bindings (e.g., `ctx.BindingTable.Bind(...)`)

The **only** mutable object exposed to an Analyzer is the `DiagnosticCollector`. Analyzers can generate:
- Diagnostics (Errors, Warnings, Info)
- Suggested Fixes
- Code Actions

But they cannot automatically alter the compilation.

### 3.3 Concurrent Execution
Because Plugins are independent and read-only, the Analyzer Engine runs them **in parallel**. All analyzers execute concurrently and only synchronize when writing to the thread-safe `DiagnosticCollector`, providing massive performance gains on multi-core hardware.

## 4. Analyzer Categories
The pluggable ecosystem supports exhaustive categorization, paving the way for a rich marketplace of enterprise rules:
- **Language Analyzers:** FDL, WDL, PMF specific idiomatic rules.
- **Domain Analyzers:** Healthcare (HIPAA), Financial (PCI/SOX, Basel), Government, ERP.
- **Style Analyzers:** Naming conventions, formatting rules, architecture patterns.
- **Security Analyzers:** OWASP rules, Secrets detection, PII enforcement.
- **Performance Analyzers:** Complexity limits, memory warnings, heavy dependencies.
- **Company Analyzers:** Internal enterprise rules (e.g., `SEC8004: Password field detected. Use PasswordHash instead.`).

## 5. The Semantic Pipeline
The full semantic evaluation flows as follows:

```text
         Parse
           │
           ▼
     Canonical AST
           │
           ▼
         Binder
           │
           ▼
    Binding Table
           │
           ▼
Core Semantic Analyzer (Pass 1...N)
           │
           ▼
Freeze SemanticModel
           │
 ┌─────────┴─────────┐
 ▼                   ▼
Analyzer 1        Analyzer 2 (Parallel Execution)
 ▼                   ▼
Diagnostics        Diagnostics
 └─────────┬─────────┘
           ▼
  Diagnostic Collector
           │
           ▼
    Code Generation
```
