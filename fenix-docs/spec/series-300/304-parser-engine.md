# SPEC-304: Parser Engine

## 1. Executive Summary
The Fénix Parser is not a single hardcoded script for FDL, WDL, or PMF. It is a universal, language-agnostic **Parser Engine** designed as a framework. By injecting specific grammar definitions, it can parse any of the Fénix DSLs while sharing 100% of the underlying parsing infrastructure (Token streams, error recovery, lookahead, and tree building).

## 2. Parsing Strategy
Fénix implements a **Hand-Written, Top-Down Recursive Descent** parser.
- **No Parser Generators:** We strictly avoid ANTLR, yacc, or PEG. Hand-written parsers are faster, easier to debug, produce superior diagnostics, and are vastly simpler to optimize for incremental compilation.
- **Predictive LL(k):** The parser avoids backtracking entirely. Decisions are made using a limited lookahead, typically `Peek(1)` or `Peek(2)`, and exceptionally `Peek(3)`.
- **Green Tree Construction:** The parser directly builds immutable `GreenNode` instances during descent. It never instantiates ephemeral `RedNode` wrappers.

## 3. Core Components

### 3.1 TokenStream
Parsers never operate directly on a raw `[]Token` slice. They interact with a `TokenStream` interface, allowing the engine to transparently inject tokens from an incremental cache in the future without modifying grammar rules.
```go
type TokenStream interface {
    Current() Token
    Peek(offset int) Token
    Next()
    Previous()
    IsEOF() bool
}
```

### 3.2 ParserContext
To avoid passing dozens of parameters through every grammar rule, the engine encapsulates the global state in a single `ParserContext`.
```go
type ParserContext struct {
    Compilation *Compilation
    Stream      TokenStream
    Diagnostics *DiagnosticCollector
    Arena       *Arena
}
```

### 3.3 Parser Base
All specific grammars (e.g., `FDLParser`, `WDLParser`) inherit from a common `ParserBase` that exposes core primitives:
- `Consume()`: Advances the stream and returns the token.
- `Match(kind)`: Consumes if the kind matches, otherwise triggers an error.
- `Expect(kind)`: Asserts the next token, synchronizes if missing.
- `Recover()`: Enters error recovery mode.
- `Synchronize()`: Skips tokens until a boundary is found.

### 3.4 NodeFactory & GreenBuilder
The Parser Engine does not instantiate Go structs manually (e.g., `&GreenEntityNode{}`). Instead, it uses a `NodeFactory` tied to a `GreenBuilder` to construct nodes.
```go
NodeFactory.NewEntity(...)
NodeFactory.NewField(...)
```
This encapsulates the logic for structural sharing, hashing (memoization), and Arena allocation (SPEC-303).

## 4. Error Recovery (Resilience)
The most critical feature of an IDE-grade parser is error resilience. The Fénix Parser Engine **never panics and never aborts** on malformed input. It always produces a structurally valid Green Tree, inserting `TokenMissing` or `TokenBad` (SPEC-301) where necessary.

### 4.1 Synchronization Sets
When the parser encounters unexpected syntax, it uses **Synchronization Sets** to recover. A synchronization set defines a list of "safe boundary" tokens (e.g., `}`, `entity`, `workflow`, `EOF`).

**Example:**
```text
entity Customer
field Name string
field Age int
<<<< ERROR >>>>
field Email string
```
If the parser hits garbage at `<<<< ERROR >>>>`, it does not fail the compilation. It loops and `Skip()`s tokens until it matches a token in the active Synchronization Set (e.g., `field` or `}`). Once found, it resumes parsing.
This allows a single compilation pass to report thousands of distinct diagnostics rather than failing at the first syntax error.
