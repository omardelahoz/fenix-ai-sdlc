# SPEC-303: Green / Red Syntax Trees

## 1. Design Goals
The Fénix Language System requires a syntax tree capable of representing millions of lines of code while supporting instantaneous incremental updates. A single mutating AST is insufficient. Following the architecture pioneered by Roslyn (C#) and Swift, Fénix implements the **Green / Red Syntax Trees** pattern.

Key design goals:
- **Immutability:** Nodes are completely immutable. Once created, they never change.
- **Structural Sharing:** Reusing identical nodes across different parts of the tree or across compilation revisions.
- **Incremental Parsing:** Rebuilding only the modified path of the tree during a keystroke edit.
- **Thread Safety:** Lock-free, fully concurrent tree traversal and analysis.
- **Lazy Materialization:** Absolute spans and parent pointers are only computed when actively requested by an IDE or the Canonicalizer.

*Note: This document defines the architectural constraints and invariants. Low-level Go memory layouts, `sync.Pool` usage, and explicit hash algorithms belong to the Implementation Notes series (`IMP-300+`).*

## 2. Why Green / Red?
A traditional AST couples semantic content with physical properties (like `Span` or `Parent`).
- If you edit line 1, all nodes below line 1 have their `Span` invalidated.
- If you share a node between two branches, its `Parent` pointer conflicts.
- Updating a parent pointer requires mutating the node or copying the entire tree.

By splitting the tree into two conceptual halves—**Green** (internal, immutable, purely structural) and **Red** (external, facade, contextual)—we achieve 100% thread safety and massive structural sharing without sacrificing the easy top-down navigation needed by semantic analyzers.

## 3. Green Tree
The Green Tree is the hidden, immutable engine of the compiler.

**GreenNode Invariants:**
- **Immutable:** Completely read-only.
- **Bottom-Up:** Constructed from the leaves (Tokens) up to the root.
- **No Parent Pointer:** A GreenNode does not know its parent.
- **Relative Width Only:** It does not know its absolute `Span`. It only knows its `Width` in bytes.
- **Hash:** It stores a deterministic `Hash` based on its kind and children.

## 4. Green Tokens
Tokens produced by the Lexer (SPEC-302) are integrated into the Green Tree as leaves. To maximize memory efficiency, the Green Tree decouples the `Token` from the `GreenNode`. The `GreenNode` simply references the token and contributes its `Width` to the parent.

## 5. Green Lists
Storing a dynamic slice of children `[]GreenNode` in every node causes millions of slice allocations, ruining GC performance. Fénix avoids this by using **Green Lists**. 
A list of elements (e.g., statements in a block) is represented as a balanced `GreenList` node rather than a Go slice. This provides `O(log N)` modification and allows structural sharing of unmodified list segments.

## 6. Red Tree
The Red Tree is the public facade used by the Semantic Analyzer, Parser, and IDE.

**RedNode Invariants:**
- **Top-Down:** Accessed starting from the root.
- **Parent Pointer:** Contains an immutable pointer to its parent `RedNode`.
- **Absolute Span:** Computes its absolute `Span` by adding its relative offset to its parent's absolute offset.
- **Ephemeral:** Only lives as long as it is needed.

## 7. Lazy Materialization
The Red Tree is never constructed completely.
When a consumer asks for `node.Children()`, the RedNode looks at its underlying GreenNode's children, wraps them in RedNodes, injects itself as their `Parent`, computes their absolute offsets, and returns them. When the consumer discards the children, they are garbage collected. The underlying Green Tree remains untouched.

## 8. Structural Sharing
Because GreenNodes lack absolute positions and parents, they are fundamentally deduplicatable.
If an entity has three fields `Field A`, `Field B`, and `Field C`:
```text
Entity
  ├── Field A (GreenNode Hash: 123)
  ├── Field B (GreenNode Hash: 456)
  └── Field C (GreenNode Hash: 789)
```
If the user edits `Field B`, the Parser creates a new `Field B'` and a new `Entity'`, but it **reuses** the exact same memory references for `Field A` and `Field C`.

## 9. Hashing & Memoization
Every GreenNode computes a `Hash` based on:
- `Kind`
- `Width`
- The `Hash` of its `Children`

Before allocating a new GreenNode, the system checks if a node with this hash already exists. This enables:
- `O(1)` equality checks (`node1 == node2`).
- Massive memory deduplication (Memoization/Flyweight).
- Lightning-fast incremental parsing.

## 10. Arena Storage
GreenNodes are allocated sequentially in an **Arena** (SPEC-301).
**Architectural Rule:** The Arena *only* stores GreenNodes. RedNodes are short-lived heap allocations (or sync.Pool wrappers) because they are just temporary navigation facades. When a compilation is discarded, the entire Arena of GreenNodes is wiped in `O(1)`.

## 11. Incremental Update
The flow for processing a keystroke edit in the IDE:
1. **Edit:** User modifies text.
2. **Lex:** Lexer scans *only* the changed region (using `EndState`).
3. **Parse:** Parser attempts to rebuild the tree.
4. **Reuse:** Parser skips unchanged nodes, pulling them directly from the previous Green Tree.
5. **New Compilation:** A new Green Root is emitted in `O(log N)` time.

## 12. Thread Safety
- **Green Tree:** 100% thread-safe. Multiple background threads (analyzers, linters) can share the exact same Green Tree memory simultaneously without locks.
- **Red Tree:** Not shared. Each thread lazily materializes its own RedNodes as it navigates down the tree.

## 13. Memory Layout
The structural footprint of a GreenNode must be radically minimized. 
Because Fénix separates `Trivia` from `Token` (SPEC-301), the Green Tree does not need to store explicit `FullWidth` vs `Width` boundaries inside every single node, unlike Roslyn.

A conceptual `GreenNode` layout only requires:
- `Kind` (uint16)
- `Flags` (uint16)
- `Width` (uint32)
- `Hash` (uint64)
- `Children` (Pointer to GreenList or Nodes)

It has:
- **No Parent Pointer**
- **No Span**
- **No URI / Filepath**
- **No Compilation Reference**
