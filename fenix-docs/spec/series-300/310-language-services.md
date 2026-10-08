# SPEC-310: Language Services Platform

## 1. Executive Summary
Fénix aims to provide world-class editor support. However, coupling the compiler directly to the Language Server Protocol (LSP, JSON-RPC) is an architectural anti-pattern. 
**SPEC-310** defines the **Fénix Language Services Platform**, a high-level, pure-Go framework that sits *above* the compilation pipeline and *below* any specific communication protocol. 

This Platform is natively consumed by the Fénix LSP Server, the CLI, and most importantly, the autonomous AI Agents of the Fénix OS that need to reason about code semantically in memory.

## 2. Platform Architecture
The Language Services Platform acts as the unified query engine for the `Workspace`.

```text
           IDE / Editor (VS Code)       AI Agents (Native Fénix)
                     │                             │
               JSON-RPC (LSP)                      │
                     │                             │
          ┌──────────▼──────────┐                  │
          │  LSP Server Wrapper │                  │
          └──────────┬──────────┘                  │
                     │                             ▼
          ┌──────────▼─────────────────────────────┴─┐
          │     Fénix Language Services Platform     │
          └──────────┬───────────────────────────────┘
                     │
          ┌──────────▼──────────┐
          │  Workspace Manager  │
          └──────────┬──────────┘
                     │
          ┌──────────▼──────────┐
          │ Compilation Snapshot│
          └─────────────────────┘
```
This architecture guarantees that an AI Agent inside Fénix can execute: `languageService.FindReferences(entityID)` natively in Go, receiving instant memory-mapped responses instead of parsing stringified JSON payloads.

## 3. The LanguageService Interface
The Platform exposes a comprehensive Go interface categorized by IDE capabilities. 

### 3.1 Architectural Principles
1. **Stateless Services:** Language Services are entirely stateless. Caching lives strictly in the `Workspace`, `Compilation`, and `SemanticModel`.
2. **Context and Cancellation:** Every method accepts a `context.Context` to allow instant cancellation when a user types a new character.
3. **Streaming:** Operations that can yield massive results (like `FindReferences`) return a channel (`<-chan Reference`) instead of a giant array, preventing memory spikes.
4. **Query-Only (Immutability):** Language Services **never** mutate a `Workspace`, `Compilation`, `SemanticModel`, `BindingTable`, `SymbolTable`, or `AST`. Any operation that conceptually modifies source code (e.g., Rename, Code Actions) returns a set of immutable `WorkspaceEdit` operations. Applying those edits is the responsibility of the editor, CLI, or AI agent.

### 3.2 The Service Contract
```go
type LanguageService interface {
    // Navigation
    GoToDefinition(ctx context.Context, pos Position) (Location, error)
    GoToImplementation(ctx context.Context, pos Position) ([]Location, error)
    FindReferences(ctx context.Context, pos Position) (<-chan Reference, error)
    
    // IntelliSense
    Completion(ctx context.Context, pos Position) ([]CompletionItem, error)
    Hover(ctx context.Context, pos Position) (HoverInfo, error)
    SignatureHelp(ctx context.Context, pos Position) (SignatureInfo, error)
    InlayHints(ctx context.Context, rng Range) ([]InlayHint, error)
    
    // Symbols & Hierarchy
    DocumentSymbols(ctx context.Context, uri DocumentURI) ([]SymbolInfo, error)
    WorkspaceSymbols(ctx context.Context, query string) (<-chan SymbolInfo, error)
    CallHierarchyIncoming(ctx context.Context, pos Position) ([]CallItem, error)
    CallHierarchyOutgoing(ctx context.Context, pos Position) ([]CallItem, error)
    CodeLens(ctx context.Context, uri DocumentURI) ([]CodeLens, error)
    
    // Refactoring & Formatting
    Rename(ctx context.Context, pos Position, newName string) (WorkspaceEdit, error)
    CodeActions(ctx context.Context, rng Range) ([]CodeAction, error)
    
    // Semantic
    SemanticTokens(ctx context.Context, uri DocumentURI) (SemanticTokens, error)
    Diagnostics(ctx context.Context, uri DocumentURI) ([]Diagnostic, error)
}
```

## 4. The Workspace Index and Event Model
To answer queries like `FindReferences` instantly without walking every file's AST, the Platform maintains a permanent **Workspace Index**.
The Index maps `NodeID -> SymbolID -> References -> Declarations -> Files`. 
It updates reactively using an **Event Model** (e.g., `WorkspaceChanged -> Invalidate -> Refresh`), ensuring the Language Service never has to "discover" changes on the fly.

## 5. O(1) Semantic Mapping (From IDE to Compiler)
How does the compiler translate a user's cursor (e.g., Line 42, Column 12) into deep semantic knowledge? It utilizes the robust mapping infrastructure established in earlier SPECs:

1. **Cursor Position:** The IDE sends `(Line 42, Col 12)`.
2. **Span Translation:** The Workspace translates this into an absolute byte offset `Span[1024..1024]`.
3. **Green Tree Traversal:** A binary search on the `GreenTree` finds the exact `RedNode` intersecting the span `O(log N)`.
4. **NodeID Retrieval:** The `RedNode` yields its unique structural `NodeID`.
5. **Binding Table Lookup:** The `SemanticModel` looks up the `NodeID` in the `BindingTable` `O(1)`.
6. **Symbol Extraction:** The binding yields a `SymbolID`.
7. **Declaration Lookup:** The `SymbolTable` yields the definition `NodeID` `O(1)`.
8. **Reverse Map to Span:** The target `NodeID` is mapped back to its `Span`, translated to Line/Col, and returned to the IDE.

## 6. Fault Tolerance and IntelliSense
While typing, the user's code is mathematically guaranteed to be syntactically broken.
Fénix Language Services are designed to **never crash** on invalid code:

- **Synchronization Sets (SPEC-304):** Ensure the Parser recovers instantly, building a partial AST populated with `MissingTokens`.
- **ErrorSymbols (SPEC-307):** Ensure the Binder doesn't panic when a type reference is half-written.
- **Synthetic Nodes (IntelliSense):** When the user types `Customer.`, the parser inserts a synthetic `IdentifierToken` representing the missing member. The Binder evaluates the type of `Customer`, queries its scope, and the Language Service returns all accessible properties of the `Customer` entity to the IDE.

## 7. AI Semantic Services and Semantic Search
Because Fénix is an AI-Native OS, it requires capabilities that JSON-RPC cannot easily express. The Language Services Platform exposes a dedicated suite of **AI Semantic Services**:

### 7.1 Semantic Search
Unlike regex-based text search, the Fénix Platform allows agents to execute deep semantic queries:
- `FindEntities()`
- `FindWorkflows()`
- `FindPipelines()`
- `FindSymbols()`

### 7.2 Native AI Endpoints
AI Agents running natively in Go can trigger heavy reasoning tasks directly on the `SemanticModel`:
- `Summarize(SymbolID)`
- `Explain(NodeID)`
- `GenerateTests(SymbolID)`
- `FindDeadCode()`
- `FindDuplicatedLogic()`
- `SuggestRefactoring()`

## 8. Incremental Sub-Millisecond Feedback
Language Services are inherently fast because they never compute anything from scratch. They strictly consume the **Immutable Compilation Snapshots** provided by the Workspace (SPEC-309). 

Because of the `Semantic Fingerprint`, if a user adds a comment or changes whitespace, the Semantic Analyzer is entirely bypassed. A request for *Semantic Tokens* or *Diagnostics* simply pulls the cached result from the previous `Compilation` in `O(1)` time, yielding sub-millisecond responsiveness in the editor.
