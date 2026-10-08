# SPEC-309: Incremental Infrastructure

## 1. Executive Summary
An AI-Native Operating System like Fénix demands an IDE that reacts in sub-milliseconds. Recompiling an entire workspace for a single keystroke is mathematically unfeasible. 

The **Incremental Infrastructure** is the orchestrator that connects the entire compilation pipeline (Lexer → Analyzer) to recycle unchanged computational work. It leverages the properties established in previous SPECs: **Immutability**, **Structural Sharing**, and **Memoization**. Fénix achieves extreme performance without relying on fragile text-diffing algorithms.

## 2. Workspace Architecture and Versioning
The architecture distinctly separates four layers of responsibility:
1. **Workspace:** Manages the overall solution, multiple projects, and cross-project references.
2. **Project:** A collection of documents and dependencies.
3. **Document:** A source file and its raw text.
4. **Compilation:** A specific, immutable snapshot of the semantic state of a Project.

To track incrementality, Fénix uses **Version Tokens**:
- `WorkspaceVersion`: Bumps when projects are added/removed.
- `ProjectVersion`: Bumps when dependencies change.
- `SyntaxVersion`: Bumps when text edits occur in a document.
- `SemanticVersion`: Bumps only when the edit alters the AST or Symbols (e.g., changing formatting updates SyntaxVersion but not SemanticVersion).

## 3. The Immutability Principle (Copy-on-Write)
Every `Compilation` object and its underlying components (Green Trees, AST, SemanticModel) are strictly **immutable**.
When a user edits a file:
1. The compiler does **not** mutate the existing `Compilation` or its AST.
2. It derives a **new** `Compilation` object (`Compilation V(n+1)`).
3. The new Compilation structurally shares up to 99.9% of its memory pointers with the previous Compilation (`Compilation V(n)`).

## 4. Hash Stability and Semantic Fingerprints (The Diff-less Architecture)
Fénix avoids tracking line-by-line text deltas from the IDE. Instead, it relies on a bottom-up hashing mechanism.

**Rule of Hash Stability:** A Structural Hash is deterministic across executions of the same compiler version and depends exclusively on canonical content. It never depends on memory addresses, allocation order, timestamps, or runtime-specific values.

Fénix maintains two types of hashes:
- **Structural Hash:** Driven by the syntax. Changes if tokens change.
- **Semantic Fingerprint:** Driven by the AST structure. (e.g., `field Name string` and `field Name \n string` have different syntax, but the exact same Semantic Fingerprint).

1. **Fast Re-Parse:** The Parser reparses the affected document and reconstructs a new Green Tree while maximizing structural reuse of unchanged subtrees from the previous Compilation.
2. **Hash Computation:** As the Canonicalizer builds the new Canonical AST, the `ASTFactory` computes the `Semantic Fingerprint` for each node.
3. **The Cut-Off Point:** If the Fingerprint of a newly parsed `Entity` matches the previous compilation, the compiler **halts** further processing for that branch. 
4. **Recycling:** It instantly recycles the previous Canonical AST subtree, its Symbols, and its Bindings.

## 4. The Incremental Pipeline

### 4.1 Incremental Green Trees
The Parser uses synchronization sets (SPEC-304) and arena allocation to rapidly build the CST. If 1 file is edited in a 10,000-file project, the new `Compilation` simply references the exact same Green Trees for the 9,999 unchanged files.

### 4.2 Incremental Canonical AST
The `ASTFactory` (SPEC-305) maintains a memoization cache keyed by structural Hashes. If a user modifies the logic inside a `Workflow`, only the nodes within that workflow, climbing up to the Root node of the file, are newly allocated. Sibling entities within the exact same file retain their original memory addresses.

### 4.3 Incremental Symbols & Bindings
Symbols are identified by an immutable `SymbolID` (SPEC-306). Changing the internal logic of a method does not alter its `SymbolID` or signature.
When the Binder (SPEC-307) detects that a Canonical AST Node's Hash is identical to the previous compilation, it skips the resolution phase entirely and maps the `NodeID` to the cached `BindingTable` entry.

### 4.4 The Dependency Graph and Cross-Project Incrementality
Fénix maintains a highly optimized **Dependency Graph** (e.g., `Entity A` depends on `Entity B`).
When a Semantic Fingerprint changes, the **Invalidation Strategy** kicks in. Fénix does not invalidate the entire project. It traverses the dependency graph and invalidates *only* the downstream consumers. 
Furthermore, this incrementality scales across projects. If `Project A` depends on `Project B`, and a public symbol in `Project B` changes its fingerprint, `Project A` is scheduled for partial recompilation.

### 4.5 Incremental Semantic Analysis
Analyzer results (Diagnostics) are cached based on the Semantic Fingerprint. If the Core Semantic Analyzer detects no changes in the semantic graph of a module, the Pluggable Analyzers (SPEC-308) are bypassed, immediately returning the previously computed Diagnostics.

## 5. IDE Integration, Scheduling, and Caching

### 5.1 Parallel Compilation Scheduler
When a keystroke triggers a new version, a **Compilation Scheduler** orchestrates the work. It identifies the dirty files and dispatches parsing and canonicalization tasks to **Parallel Workers**. Once the workers finish, the Scheduler merges the results into the new `Compilation`.

### 5.2 Fast Cancellation
In an IDE, a user might type 10 characters per second. Triggering 10 full compilations is wasteful. Every pass in the incremental infrastructure accepts a `context.Context` (or `CancellationToken`). The moment a new edit arrives, the Scheduler cancels the in-flight `Compilation(n)`, immediately freeing CPU resources to begin `Compilation(n+1)`.

### 5.3 Cache Policy and Garbage Collection
Fénix defines strict cache lifecycles to prevent memory leaks:
- **Workspace Cache:** Lives for the duration of the IDE session (e.g., Project Graph).
- **Compilation Cache:** Lives only while active (LSP requests).
- **AST / Binding Cache:** Managed by the `ASTFactory` with an eviction policy (e.g., LRU).

Once an old `Compilation` snapshot is no longer needed by any background LSP thread or build process, it is released. The Go Garbage Collector reclaims only the nodes that were modified and orphaned, leaving the massive shared structure intact.
