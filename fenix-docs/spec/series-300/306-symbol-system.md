# SPEC-306: Symbol System

## 1. Executive Summary
The Canonical AST (SPEC-305) defines the structure and intent of the source code. However, it lacks true semantic identity. When the AST says `Customer`, it is simply a string token inside a `TypeReference` node. 

The **Symbol System** bridges this gap by providing an independent, strongly typed semantic graph where `Customer` is a uniquely resolved entity with properties, scope, and identity. This system unifies the semantic models of FDL, WDL, and PMF.

## 2. Syntax vs. Semantics
It is critical to distinguish between Syntax and Semantics:
- **AST (Syntax):** Answers *how* the code was written. It contains syntax structures, expressions, and literals.
- **Symbols (Semantics):** Answers *what* the code represents. It contains identity, type information, and scoping rules.

A single Symbol may be backed by multiple AST declarations (e.g., `partial entity Customer` spread across two files), proving that Symbols and AST nodes do not have a 1:1 relationship.

## 3. Symbol Taxonomy
Symbols form a strongly typed hierarchy that represents all concepts across the Fénix DSL ecosystem.

### 3.1 Base Symbol Interface
Every symbol implements a core interface ensuring universal traversability. Crucially, a Symbol is defined by its immutable identity (`SymbolID`), not its name.
```go
type SymbolID uint64

type Symbol interface {
    ID() SymbolID
    Name() string
    Kind() SymbolKind
    ContainingSymbol() Symbol    // The logical parent (e.g., Namespace or Entity)
    Declarations() []Declaration // Links back to the AST (via NodeID)
}
```

### 3.2 Expanded Symbol Taxonomy
The hierarchy provides exhaustive coverage for all concepts across FDL, WDL, and PMF. To ensure extensibility, the architecture anticipates the following symbol kinds:
- **Containers:** `Namespace`, `Module`, `Import`, `Alias`
- **Types:** `PrimitiveType`, `ArrayType`, `GenericType`
- **FDL (Data):** `Entity`, `Property`, `Enum`, `EnumMember`, `Service`, `Operation`
- **WDL (Logic):** `Workflow`, `Activity`, `Parameter`, `Variable`
- **PMF (Infra):** `Pipeline`, `Stage`

## 4. The Namespace Graph and SymbolTable

### 4.1 The Namespace Graph
`NamespaceSymbol` is not just another symbol; it is the backbone of the semantic tree. It organizes symbols hierarchically (Global -> Company -> Project -> Module -> Types):
```text
Global Namespace
   ├── com
   │    └── enterprise
   │         ├── crm
   │         │    ├── Customer (EntitySymbol)
   │         │    └── Order (EntitySymbol)
```
This guarantees that multiple DSLs can safely coexist without naming collisions.

### 4.2 The SymbolTable
The `SymbolTable` is attached to the `Compilation` and acts as the central registry. It provides `O(1)` lookups for the Binder and the LSP:
```go
type SymbolTable struct {
    GlobalNamespace *NamespaceSymbol
    ByID            map[SymbolID]Symbol
    ByNode          map[NodeID]Symbol
}
```

## 5. Declarations vs. Symbols (The Missing Link)
Symbols do not emerge directly from the AST. There is a critical intermediate layer: **Declarations**.
A `Declaration Collector` iterates the Canonical AST and extracts lightweight references, which are then passed to a `SymbolFactory`. 
For example, a user might write `partial entity Customer` in three different files. The collector extracts 3 `Declarations`, but the factory fuses them into exactly 1 `EntitySymbol`.

To preserve memory (SPEC-301), Symbols **never hold pointers to AST nodes**. They hold `Declarations`:
```go
type Declaration struct {
    NodeID   NodeID
    Location Span
}
```
Some languages distinguish between declarations and definitions. In Fénix, a Symbol acts as the ultimate definition, aggregating all of its constituent declarations.

## 6. Symbol Factories
To maintain architectural purity, the Binder never manually instantiates Go structs (e.g., `&EntitySymbol{}`). All symbol creation passes through a `SymbolFactory` attached to the Compilation:
```go
ctx.Compilation.SymbolFactory.NewEntity(...)
```
This centralizes the assignment of `SymbolID`s and registration in the `SymbolTable`.

## 7. Symbol Lifecycle and Architecture
The flow from text to pure semantics is strictly unidirectional:
```text
Source File -> Lexer -> Parser -> Canonicalizer -> Canonical AST
   │
   ▼
Declaration Collector (Hoists structures from AST)
   │
   ▼
Symbol Factory (Assigns ID, creates Symbol)
   │
   ▼
Symbol Table (Registers Symbol globally/by namespace)
   │
   ▼
Binder (Resolves cross-references)
   │
   ▼
Freeze (Symbols become 100% Immutable)
   │
   ▼
Semantic Analyzer (Consumes frozen Symbols)
```

## 8. Symbol Invariants and Equality
To ensure correctness across caches, incremental compilations, and the LSP, the Symbol System adheres to strict invariants:

### 8.1 Symbol Identity (Equality)
Two symbols are equal **if and only if** they possess the exact same `SymbolID`. 
Equality never depends on the symbol's name, file, AST declaration, or namespace. This makes structural renaming or cache invalidation trivial.

### 8.2 ContainingSymbol Rules
The `ContainingSymbol` forms a rigid tree:
- It never changes after the symbol is created.
- It is never circular.
- It is never `nil`, **except** for the `GlobalNamespace`.

### 8.3 Core Architectural Invariants
1. **Symbols are immutable** after binding.
2. **Symbols never reference CST nodes.**
3. **Symbols never reference RedNodes.**
4. **Symbols may reference other Symbols.**
5. **Symbols are 100% thread-safe.**
6. **Symbols belong to exactly one Compilation.**
7. **Symbols never mutate after publication.**
