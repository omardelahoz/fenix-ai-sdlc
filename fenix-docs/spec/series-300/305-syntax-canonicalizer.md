# SPEC-305: Syntax Canonicalizer

## 1. Executive Summary
The Fénix **Syntax Canonicalizer** bridges the gap between raw, messy human input (CST) and the pure, deterministic structural representation (Canonical AST) required for semantic analysis. 

The Canonicalizer is **not** a simple Visitor pattern applied to the CST. It is a robust, lossy transformation pipeline that desugars, normalizes, and isolates meaning from physical syntax.

## 2. Enriched SyntaxTree & Lossy Transformation
The Canonicalizer enriches the existing `SyntaxTree` by producing its immutable Canonical AST. The CST remains unchanged, and both views coexist within the same `SyntaxTree`.

However, the Canonicalization process is intentionally a **lossy transformation**. It strips away information at each step to reach a pure state:

```text
Source Text
    │
    ▼
Lexer (Tokens)
    │
    ▼
Green Tree (Structural Sharing, 100% Fidelity)
    │
    ▼
Lazy Red Tree (Navigation Facade)
    │
    ▼
Syntax Canonicalizer (Lossy Transformation)
    │
    ▼
Canonical AST (Pure Meaning)
```

## 3. Canonicalization Context
To avoid passing numerous parameters, the process operates within a context:
```go
type CanonicalizationContext struct {
    Compilation *Compilation
    SyntaxTree  *SyntaxTree
    Diagnostics *DiagnosticCollector
}
// Note: Factories (like ASTFactory) are accessed via ctx.Compilation.ASTFactory
```

## 4. The Canonical AST
The output of the Canonicalizer is strictly the **Canonical AST**.

**Architectural Independence:** The AST represents pure semantic intent. It has absolutely no knowledge of the parser's internal mechanics. The AST **never** contains:
- Tokens
- Trivia (whitespace, regular comments are discarded)
- GreenNodes or RedNodes
- ParserContext or Synchronization info
- Missing tokens or recovery garbage
- Parent Pointers

**No Parent Pointers:** The Canonical AST intentionally omits parent pointers to maximize structural sharing, immutability, and parallel traversal. Contextual navigation is provided through a `VisitorContext` rather than embedded references.

*(Note: Documentation comments (`doc-comments`) survive because they are promoted from lexical trivia into semantic metadata. They contribute to the public semantic model consumed by documentation generators, IDEs, and language services).*

### 4.1 Canonical AST Equivalence
Two Canonical AST nodes are equivalent if their `Kind`, `Children`, and semantic `Properties` match. It does not matter if the original source had different comments, spacing, line breaks, or error recovery nodes.

AST nodes may expose an internal structural hash used by memoization and incremental compilation.

### 4.2 Normalization vs Desugaring
These are two distinct transformations within the Canonicalizer:
- **Normalization:** Squashes different syntactic ways of expressing the same semantic concept. E.g., `Customer?` and `Nullable<Customer>` both become `TypeReference(Nullable=true)`.
- **Desugaring:** Breaks down syntax sugar into primitive semantic nodes. E.g., `field Age int = 10` becomes a `Field` containing an `Initializer` and a `LiteralExpression`.

### 4.3 Attributes Attachment
Annotations and attributes (e.g., `@required`, `@length(10)`) are physically separate nodes in the CST. The Canonicalizer groups and attaches attributes **preserving source order** to the corresponding `Field` or `Entity` in the AST.

### 4.4 Syntactic Validation
The Canonicalizer may emit diagnostics when malformed CST constructs cannot be represented in the Canonical AST (e.g., detecting `@required 123 entity`). It does not replace the Parser's syntactic validation, but catches structural impossibilities before semantic analysis.

## 5. Canonicalization Rules & Invariants
Canonical AST nodes are **immutable value objects**. Equality is structural rather than referential. Every transformation must satisfy being:
- Deterministic
- Idempotent (canonicalizing an already canonicalized AST yields the exact same AST)
- Side-effect free
- Independent of source formatting
- Independent of source file location

The Canonicalizer also guarantees:
1. **Never modifies the Green Tree.**
2. **Never modifies the Red Tree.**
3. **Always produces a valid Canonical AST.**
4. **Never reuses RedNodes.**
5. **Never contains Trivia, Tokens, or Recovery nodes.**

## 6. AST Factory & Memoization
The Canonicalizer does not instantiate Go structs manually. It utilizes an `ASTFactory` (accessed via `ctx.Compilation.ASTFactory`) that receives the context, allowing for future extensions (memo caches, arenas):
```go
ASTFactory.NewEntity(ctx, ...)
ASTFactory.NewField(ctx, ...)
ASTFactory.NewExpression(ctx, ...)
```

### 6.1 Incremental Canonicalization (Structural Sharing)
The AST supports massive deduplication. If two branches generate the exact same structure, the `ASTFactory` reuses the exact same AST node via memoization.

*Note: Memoization is optional. Implementations may disable AST memoization if memory usage outweighs the expected benefits. The Canonical AST remains functionally identical regardless of memoization.*

If a user edits a single property `B` in an entity:
```text
Entity
 ├── A
 ├── B  <-- EDITED
 └── C
```
The Canonicalizer does not rebuild `A` or `C`. It only canonicalizes `B'`, and links the existing `A` and `C` to the new `Entity'`.

## 7. Canonicalization Pipeline Stages
The Canonicalizer is not a single algorithm, but a sequence of well-defined transformations:
```text
      Red Tree
         │
         ▼
   Normalization
         │
         ▼
  Type Normalization
         │
         ▼
    Desugaring
         │
         ▼
Documentation Extraction
         │
         ▼
 Attribute Attachment
         │
         ▼
  AST Memoization
         │
         ▼
    Canonical AST
```
