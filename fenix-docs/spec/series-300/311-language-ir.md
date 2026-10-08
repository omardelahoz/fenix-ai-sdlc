# SPEC-311: Canonical Language IR

## 1. Executive Summary
With the semantic pipeline complete (SPEC-308) and query services operational (SPEC-310), the compiler must translate Fénix code into an executable format. The Fénix Language spans multiple DSLs like FDL (Entities), WDL (Workflows), and PMF (Agents). Compiling these directly to a target language (like Go or TypeScript) from the AST is exponentially complex. 

**SPEC-311** introduces the **Canonical Language IR (Intermediate Representation)**. It is a High-Level IR (HIR), inspired by Roslyn's `IOperation`, designed to normalize the semantic model into a purely logical tree. It strips away syntax, trivia, and syntax sugar, making code generation completely backend-agnostic.

## 2. The High-Level IR (HIR) Architecture
Unlike a Low-Level SSA (Single Static Assignment) graph used in LLVM for binary compilation, Fénix utilizes a **High-Level IR (HIR)**. 
- **Tree-like Structure:** It retains a hierarchical structure but represents *operations* and *semantics* rather than *syntax*.
- **Syntax Independent:** An `IfStatement` in FDL and a `ConditionalAction` in WDL might parse differently into the AST, but they both lower to the exact same `IConditionalOperation` in the IR.
- **Backend Agnostic:** The IR contains no target-specific constructs. A Code Generator can consume the Fénix IR to emit Go, TypeScript, Python, or raw YAML definitions for the runtime engine.

## 3. The Lowering Process (Desugaring)
The transformation to IR is driven by the semantic meaning of the code, not just its syntax. Therefore, Lowering strictly depends on the `SemanticModel`.

The pipeline is:
`Canonical AST -> Binder -> SemanticModel -> Lowering -> Canonical IR -> Optimization -> Emitter`

During Lowering, the compiler:
1. **Strips Trivia and Spans:** The IR does not care about source code formatting, whitespace, or file offsets.
2. **Desugars Syntax:** Complex language features (like macros or implicit loops) are rewritten into simpler primitive operations.
3. **Embeds Semantic Identity:** Unlike the AST, every IR node inherently encapsulates its resolved `SymbolID` and `TypeSymbol`. There is no need to query a separate Binding Table; the IR is fully self-contained.

Example:
```text
// Original FDL
entity Customer {
    property Name string?
}
```
Lowers into an IR graph of:
```text
IEntityDeclaration (SymbolID: Customer)
└── IPropertyDeclaration (SymbolID: Name, Type: Nullable<String>)
```

## 4. The IOperation Interface
Every node in the IR implements a base interface representing a semantic operation:
```go
type OperationKind int

type IOperation interface {
    Kind() OperationKind
    Type() SymbolID           // The resulting type of this operation (e.g., Int, Void)
    ConstantValue() any       // If the operation evaluates to a constant at compile time
    SyntaxNode() NodeID       // Pointer back to the AST (for generating source maps/errors)
    Children() []IOperation
}
```

Specific operations extend this base:
```go
type IInvocationOperation interface {
    IOperation
    TargetMethod() SymbolID
    Arguments() []IArgumentOperation
}

type IConditionalOperation interface {
    IOperation
    Condition() IOperation
    TrueBranch() IOperation
    FalseBranch() IOperation
}
```

## 5. Optimization Passes
Because the IR is a mathematically pure representation of intent, Fénix can perform compiler-level optimizations *before* generating code:
- **Constant Folding:** `3 + 4` becomes `7`.
- **Dead Code Elimination:** Removing `if (false) { ... }` blocks.
- **Workflow Inlining:** Flattening WDL steps for performance.

These passes transform the IR tree, producing a final optimized IR ready for the Code Generation backends.

## 6. Code Generation Pipeline (Emitters)
Because the IR is perfectly normalized and optimized, writing a code generator (Emitter) is a highly streamlined task. 
Fénix does not write directly to the file system. It generates intermediate documents that are formatted before writing.

```text
       Canonical Language IR
                 │
   ┌─────────────┼─────────────┐
   ▼             ▼             ▼
 GoEmitter    TsEmitter    YamlEmitter
   │             │             │
   ▼             ▼             ▼
  GenDoc        GenDoc        GenDoc
   │             │             │
   ▼             ▼             ▼
Formatter     Formatter     Formatter
   │             │             │
   ▼             ▼             ▼
FileSystem    FileSystem    FileSystem
```

An `Emitter` is a Visitor over the `IOperation` tree that produces a `GeneratedDocument`. This allows Fénix to effortlessly support outputting Go, TypeScript, YAML, JSON, Markdown, HTML, or OpenAPI specs through a unified, formatting-aware pipeline.
