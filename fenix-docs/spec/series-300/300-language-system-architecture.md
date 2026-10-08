# SPEC-300: Language System Architecture

## 1. Executive Summary
Fénix defines three Domain-Specific Languages (DSLs): FDL (product definition), WDL (workflow orchestration), and PMF (processor manifests). Rather than building three independent parsers, Fénix implements a **unified Language System** modeled after industrial compiler architectures (LLVM, Roslyn, TypeScript).
All three languages share a common compilation pipeline, common data structures, and a common diagnostics engine. Only the grammar rules and semantic validation logic differ per language.

## 2. Syntax Philosophy
Fénix DSLs are **declarative languages for describing systems**, not imperative programming languages. Their syntax follows these principles:

- **Declarative, not imperative.** Inspired by HCL (Terraform) and CUE. No loops, no conditionals, no side effects.
- **No YAML.** YAML's indentation sensitivity, weak typing, and poor error messages are incompatible with a serious engineering tool.
- **No C-style braces for structure.** FDL describes *what* a system is, not *how* it runs. It should read like a specification, not source code.
- **LL(1) Grammar.** Recursive Descent parsing. Minimal lookahead. Zero ambiguity. Easy to maintain, easy to extend.

## 3. The Workspace Compilation Model (Roslyn Model)
Fénix does not compile isolated files. It compiles entire projects within a workspace. The compilation model is hierarchical:

```text
Workspace
    │
    ▼
Project (e.g., Fénix backend, Fénix UI)
    │
    ▼
Compilation (The immutable state of the compiler for a project)
    │
    ├─► SyntaxTrees (CSTs and ASTs from all source files)
    ├─► SemanticModels (Symbols, BindingTable, diagnostics)
    └─► CanonicalLanguageIR (The pure domain representation)
```

**Crucial Architecture Note:** The `SemanticModel` depends on the `Compilation`, not vice versa. A single `Compilation` can spawn multiple `SemanticModels` (e.g., one for Release, one for Analyzer Plugins) operating over the same `SyntaxTrees`.

## 4. The Compiler Pipeline
Every Fénix DSL source file passes through the following deterministic, multi-phase pipeline within a `Compilation`.

```text
Source Text
    │
    ▼
┌──────────────────┐
│  Lexer           │   Language-agnostic. Emits generic tokens + Trivia.
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Parser Engine & │   The Engine handles tree building & error recovery.
│  Grammar Provider│   The Grammar defines rules (e.g., FDL vs WDL).
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  CST             │   Concrete Syntax Tree. Preserves ALL source text.
│  (Green/Red Tree)│   Designed with immutable Green nodes and Facade Red nodes.
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Syntax          │   Canonicalization engine. Expands defaults, resolves
│  Canonicalizer   │   aliases, orders attributes. Produces the canonical AST.
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  AST             │   Abstract Syntax Tree. Pure semantics. Immutable.
│  (Canonical)     │   NO parent pointers.
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Binding Engine  │   Resolves string references ("Customer") to actual
│  (Binder)        │   Symbols. Produces a BindingTable (NodeID -> Binding)
│                  │   without duplicating the AST.
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Semantic        │   Operates on the BindingTable to validate rules,
│  Analyzer        │   constraints, and complex type relations.
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Canonical       │   Source-agnostic domain model representing "WHAT exists"
│  Language IR     │   (e.g., IREntity, IRService). Decoupled from AST.
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Engineering IR  │   The final physical realization of the system representing
│  Builder         │   "HOW it is built" (e.g., Postgres Table, REST Endpoint).
└──────────────────┘
```

## 5. Parser Engine vs. Grammar Separation
There is no "FDL Parser". There is a single, robust **Parser Engine** (an LL(1) / Pratt parsing framework) that understands tokens, error recovery, and CST construction. 
The specific rules for FDL, WDL, and PMF are injected via a `Grammar` interface. This allows Fénix to easily add a `PlantUMLGrammar` or `OpenAPIGrammar` in the future without rewriting parsing logic.

## 6. CST vs AST & The Green/Red Tree Pattern
### Concrete Syntax Tree (CST)
The CST is a **full-fidelity** representation of the source text. It preserves every token, every piece of Trivia (whitespace/comments), and exact formatting.
It uses Roslyn's **Green Tree / Red Tree** pattern (SPEC-303):
- **Green Tree:** Purely immutable, hashable, shares memory across identical subtrees, and contains no parent pointers or absolute spans.
- **Red Tree:** An ephemeral wrapper (facade) created on-demand. It provides absolute `Span`, `Parent` pointers, and navigation for the IDE.

### Abstract Syntax Tree (AST)
The AST is a **semantic-only** representation. It discards formatting Trivia but preserves doc-comments.
**Architectural Rule:** The AST is purely immutable and **has no `Parent()` pointers**. Parent traversal is strictly prohibited in the AST to enable perfect structural sharing, cacheability, and parallel visitors. Ascending the tree is managed by the `VisitorContext`.

## 7. The Binding Engine and BindingTable
Because Fénix DSLs are declarative, they lack complex imperative expressions (Control Flow Graphs, lowering, etc.). Creating a whole new `BoundTree` (a common pattern in general-purpose languages) would needlessly waste memory and GC cycles.
Instead, the **Binder** resolves string references (e.g., `"Customer"`) into `Bindings` and stores these mappings in a **BindingTable** attached to the `SemanticModel`. The Semantic Analyzer traverses the AST and retrieves pre-resolved `Bindings` directly from the BindingTable.

## 8. The Symbol System (SPEC-306)
Symbols are first-class citizens. The Binder produces Symbols (`EntitySymbol`, `TypeSymbol`, `WorkflowSymbol`), and the Semantic Analyzer and IR Builders consume Symbols. This cleanly decouples syntax from semantic identity.

## 9. The Syntax Canonicalizer
Responsibilities:
- **Strip formatting Trivia** (preserving doc-comments).
- **Expand defaults:** (e.g., inject `String` if a field lacks a type).
- **Resolve aliases:** Normalize shorthand syntax.
- **Order attributes:** Sort `@` attributes deterministically.
- **Canonicalize expressions:** Standardize attribute argument formats.

## 10. Canonical Language IR vs. Engineering IR
A core tenet of the Fénix Language System is the hard boundary between the syntax representation and the output.
- **Canonical Language IR (WHAT exists):** A pure domain model. It defines logical entities, workflows, and services. **Absolute Rule:** It MUST NOT import, reference, or contain `ASTNode`, `CSTNode`, or `Span`. This purity allows generators like OpenAPI or Terraform to feed directly into the Canonical Language IR.
- **Engineering IR (HOW it is built):** Represents physical, deployable artifacts (e.g., `Postgres Table`, `Kafka Topic`, `Angular Form`). Defined by the Fénix Runtime (SPEC-102).

## 11. Incremental Parsing & Memory (Architectural Reservation)
The architecture reserves the capability for sub-millisecond incremental updates via:
- **Green/Red Trees:** Subtrees can be structurally shared across edits.
- **Stable Node IDs:** Every node has a stable identifier based on `StablePath` (NOT absolute Span, which ruins incremental caching on every newline).
- **Arena Allocation:** All compiler nodes are allocated in an Arena to eliminate Garbage Collector pressure.

## 12. LSP Readiness (Architectural Reservation)
The Language System is designed as a pure, stateless backend for a Language Server (LSP). The server itself will simply be a thin JSON-RPC adapter wrapping this infrastructure.
