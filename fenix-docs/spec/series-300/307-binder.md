# SPEC-307: Binder

## 1. Executive Summary
With the **Canonical AST** providing structural intent and the **Symbol System** providing identity, the compiler requires an engine to connect the two. The **Binder** is the semantic bridge of Fénix. 

It traverses the AST, evaluates contextual references (e.g., `Customer` inside a `TypeReference`), and securely links the syntax node to the `EntitySymbol`. Crucially, the Binder is non-destructive: it **never** mutates the AST and **never** mutates the Symbols.

## 2. The Semantic Binding Graph
The sole output of the Binder is the `BindingTable` (stored inside the `SemanticModel`), which conceptually forms a rich **Semantic Binding Graph**.

Instead of adding a `Symbol` pointer directly to an `ASTNode` (which would break immutability), the Binder produces external linkages. Furthermore, to maximize cache-friendliness, serialization, and incremental compilation, bindings store `SymbolID`s rather than memory pointers:
```go
type BindingKind int
const (
    TypeBinding BindingKind = iota
    PropertyBinding
    NamespaceBinding
    VariableBinding
    ConstantBinding
)

type Binding struct {
    Kind          BindingKind
    SymbolID      SymbolID
    TypeID        SymbolID
    ConstantValue ConstantValue
    Diagnostics   []Diagnostic
}

type BindingTable struct {
    Bindings map[NodeID]*Binding
}
```
All bindings are created through a `BindingFactory` (e.g., `ctx.BindingFactory.NewTypeBinding(...)`) rather than manual struct initialization.

## 3. Binding Context
To avoid passing excessive parameters, the Binder operates within a context:
```go
type BindingContext struct {
    Compilation    *Compilation
    SemanticModel  *SemanticModel
    SymbolTable    *SymbolTable
    BindingTable   *BindingTable
    BindingFactory *BindingFactory
    Diagnostics    *DiagnosticCollector
}
```

## 4. Two-Pass Resolution
Because declarative languages permit circular dependencies and forward references (using a type before it is declared), the Binder operates in two conceptual passes:

### Phase 1: Declaration Hoisting (SPEC-306)
Before any resolution occurs, the `Declaration Collector` scans the Canonical AST, extracts `Declarations`, and uses the `SymbolFactory` to populate the `SymbolTable`. By the end of this phase, the global namespace knows every entity, workflow, and property that exists in the `Compilation`.

### Phase 2: Semantic Resolution (Binding)
The Binder evaluates expressions and type references against the fully populated `SymbolTable`. However, the Binder itself does not search the table directly. It delegates this to a specialized `NameResolver`.

For a `Field` declaration `Age int`, the Binder:
1. Looks at the type reference `int`.
2. Asks the `NameResolver` to resolve `int`.
3. The `NameResolver` checks the active scope, imported namespaces, referenced projects, and the Standard Library.
4. It finds the primitive `Int` `SymbolID`.
5. The `BindingFactory` writes a new entry `NodeID -> Binding{TypeID: Int}` into the `BindingTable`.

## 5. Lazy Evaluation and Thread Safety
Fénix is designed for massive enterprise codebases. Resolving every single reference across thousands of files simultaneously (Eager Binding) would cause severe CPU and memory spikes.

Following the architecture of modern compilers (Roslyn, TypeScript), the Fénix Binder is **Lazy**. 
During Phase 2, the `BindingTable` starts empty. Bindings are resolved strictly **on-demand**:
1. `SemanticModel.GetBinding(NodeID)` is called.
2. If the binding exists in the cache, it returns immediately.
3. If not, it computes the binding, caches it in the `BindingTable`, and returns it.

Once calculated, a binding is **never** recalculated.

### 5.1 Thread Safety
Because multiple concurrent semantic analyzers or LSP requests may ask for `Binding(Node42)` simultaneously, the computation must guarantee that **multiple concurrent requests produce exactly one Binding**. The architecture guarantees thread-safe, single-flight computation (e.g., via `sync.Once` or compare-and-swap) without requiring the consumer to manage locks.

## 6. Cross-Project Scope Resolution Hierarchy
The `NameResolver` resolves names by searching outward through the `ContainingSymbol` hierarchy, scaling across boundaries:
1. **Local Scope:** (e.g., variables inside a workflow block).
2. **Member Scope:** (e.g., sibling properties inside an Entity).
3. **Current Namespace:** (e.g., `com.enterprise.crm`).
4. **Imported Namespaces:** (e.g., explicit imports).
5. **Referenced Projects:** (e.g., shared company libraries).
6. **Standard Library / Global Namespace:** Fallback for built-in primitives.

## 7. Resilience and Error Recovery
Like the Parser Engine, the Binder is built for resilience. It **never** panics or aborts compilation upon encountering an unresolved reference.
If a user types `Customer` but the symbol does not exist:
1. The `NameResolver` emits a `Diagnostic` (`SymbolNotFoundError`).
2. It yields an `ErrorSymbol` (representing an unknown type, namespace, or property).
3. The Binder stores the `Binding` with the `ErrorSymbol`.
This allows downstream analyzers to continue operating smoothly without crashing due to `nil` pointers.

## 8. Separation of Concerns: No Validation
It is critical to note what the Binder **does not** do. The Binder only answers questions of identity:
- *What symbol is this?*
- *What type does this resolve to?*

The Binder **does not** validate business or semantic rules. It never asks:
- *Is this type assignment valid?*
- *Does this workflow have a circular dependency?*
- *Is this property allowed to be nullable?*
Rule validation is strictly the responsibility of the **Semantic Analyzer** (SPEC-308).

## 9. Binding Lifecycle
The binding phase perfectly caps the semantic preparation before analysis begins:
```text
      Canonical AST
             │
             ▼
   Declaration Collector
             │
             ▼
       Symbol Factory
             │
             ▼
        Symbol Table
             │
             ▼
           Binder
             │
             ▼
       Binding Table
             │
             ▼
           Freeze
             │
             ▼
     Semantic Analyzer
```
