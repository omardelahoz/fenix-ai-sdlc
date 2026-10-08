# SPEC-301: Common Language Infrastructure

## 1. Executive Summary
Before defining any grammar (FDL, WDL, or PMF), Fénix must establish the shared primitive types that every language in the system will use. This specification defines the **Common Language Infrastructure** — the foundational Go types for tokenization, Trivia handling, CST/AST node models, rich type references, the Symbol System (SPEC-306), the Binding Graph, structured diagnostics with stable error codes, memory arenas, and the workspace compilation model.
All types defined here live in `/core/internal/parser/common/` and are imported by every language-specific parser.

## 2. Workspace Compilation Model
Fénix compiles projects, not isolated files. The compiler's state is encapsulated in a `Compilation` object (following the Roslyn model).

```go
type Workspace struct {
    Projects map[string]*Project
}

type Project struct {
    Name    string
    Sources []*SourceFile
}

type Compilation struct {
    Project     *Project
    SyntaxTrees []*SyntaxTree       // Contains both CST and AST
    Version     uint64              // For incremental cache invalidation
}

// SyntaxTree preserves the 1:1 relationship between a concrete and canonical view.
type SyntaxTree struct {
    CST *RedTree    // Pointer to the Red Tree root (facade)
    AST ASTNode     // Pointer to the canonical AST root
}

// SemanticModel sits on top of a Compilation and resolves symbols and bindings.
// Multiple semantic models can exist for a single Compilation (e.g., Debug, Release).
type SemanticModel struct {
    Compilation  *Compilation
    Symbols      *SymbolTable
    BindingTable *BindingTable       // Replaces BoundTree to save memory
    Diagnostics  *DiagnosticCollector
}
```

### Source File
The `SourceFile` is the top-level container for a single file's text and mapping.
```go
type SourceFile struct {
    URI       string     // Absolute file path or virtual URI
    Text      []byte     // Raw UTF-8 source content
    Encoding  string     // e.g., "utf-8"
    SourceMap *SourceMap // Line/column lookup table
}
```

## 3. Arena Allocation & Node Identity
Compiler nodes (Tokens, CST, AST, Types) are instantiated by the thousands. To eliminate Garbage Collector (GC) pressure, Fénix allocates all language nodes inside an **Arena**. When a `Compilation` or `SemanticModel` is discarded, its Arena is wiped instantly.

### NodeID (Stable Identity without Spans)
If `NodeID` includes the `Span` (absolute file offsets), inserting a single newline invalidates all downstream IDs, destroying semantic caches. Instead, `NodeID` is derived from a `StablePath`.

**StablePath Specification**
A `StablePath` represents the lineage of a node relative to the root, expressed as a sequence of child indices. It is deterministic within a syntax tree revision but is not intended to survive structural edits. For example:
```text
Root
 ├── Entity[0]
 │      ├── Field[0]
 │      └── Field[1]
 └── Entity[1]
```
The paths would be `0`, `0/0`, `0/1`, `1`.

```go
// NodeID is resilient to text shifting.
type NodeID string // Calculated hash of: StablePath (kind + stablepath)
```

## 4. Two-Layer Lexer Primitives
The Fénix Lexer is **language-agnostic**. It does not recognize keywords. It emits generic tokens and attaches Trivia (whitespace, comments) to significant tokens.

### Position & Span
```go
type Position struct {
    Offset int
    Line   int
    Column int
}

type Span struct {
    Start Position
    End   Position
    File  string
}
```

### Lexer State and Result
```go
type LexerState uint8

const (
    LexerDefault LexerState = iota
    LexerInString
    LexerInMultilineString
    LexerInInterpolation
)

type LexResult struct {
    Tokens      []Token
    Diagnostics []Diagnostic
    EndState    LexerState
}
```

### Trivia
```go
type TriviaKind int

const (
    TriviaWhitespace TriviaKind = iota
    TriviaNewline
    TriviaComment
    TriviaDocComment
)

type Trivia struct {
    Kind TriviaKind
    Span Span
}
```

### Token
```go
// Token represents a pure, lightweight lexical element.
type Token struct {
    ID             NodeID
    Kind           TokenKind
    Span           Span
}

// TokenInfo wraps a Token with its associated Trivia when formatting or fidelity is needed.
type TokenInfo struct {
    Token
    LeadingTrivia  []Trivia
    TrailingTrivia []Trivia
}
```

### Generic TokenKind
Keyword resolution is deferred to the Parser Engine.
```go
type TokenKind int

const (
    TokenIdentifier     TokenKind = iota
    TokenStringLiteral
    TokenIntLiteral
    TokenFloatLiteral
    TokenSymbol                            // { } ( ) . , :
    TokenOperator                          // = >= <= *
    TokenAt                                // @
    TokenDocComment                        // ///
    TokenComment                           // //
    TokenNewline
    TokenWhitespace
    TokenEOF
    TokenBad                               // Unrecognized lexical garbage (e.g. %)
    TokenMissing                           // Synthetic token inserted by parser for error recovery
)
```

## 5. Concrete Syntax Tree (CST) Node Model
The CST preserves the **full fidelity** of the source text. It uses the **Green/Red Tree** pattern (SPEC-303).

### CST Node Interface (Red Facade)
```go
type CSTNode interface {
    ID()        NodeID
    Kind()      CSTNodeKind
    Span()      Span         // Calculated absolute span
    Children()  []CSTNode
    Tokens()    []Token
    Parent()    CSTNode      // Pointer to parent (Red tree only)
    IsError()   bool
}
```

## 6. Abstract Syntax Tree (AST) Node Model
The AST is the **semantic-only** representation produced by the Syntax Canonicalizer.
**Crucial Architecture Rule:** The AST is purely immutable and lacks `Parent()` pointers. This allows massive parallelism, memoization, and structural sharing.

### AST Node Interface
```go
type ASTNode interface {
    ID()          NodeID
    Kind()        ASTNodeKind
    Span()        Span
    Children()    []ASTNode
    DocComment()  string            // Extracted /// comment
    Attributes()  []Attribute       // @name(args) metadata
    Accept(Visitor)
}
```

### Attribute (Rich, Typed Arguments)
```go
type Attribute struct {
    ID        NodeID
    Name      string
    Arguments []AttributeArgument
    Span      Span
}

type AttributeArgument struct {
    Name  string
    Value Expression   // Typed value
    Span  Span
}
```

### Expression (Attribute Value Types)
```go
type Expression interface {
    exprNode()
    Span() Span
}
// Implementations: StringExpr, IntExpr, FloatExpr, BoolExpr, ListExpr, IdentExpr
```

### TypeReference
Type annotations support generics and namespaces.
```go
type TypeReference struct {
    ID         NodeID
    Name       string            // e.g., "Customer", "List"
    Namespace  string            // e.g., "Booking.API"
    TypeArgs   []TypeReference   // e.g., [Customer] for List<Customer>
    Nullable   bool
    Span       Span
}
```

### Visitor Pattern (With Context)
Because `ASTNode` has no `Parent()` pointer, upward navigation and contextual awareness are managed entirely by the `VisitorContext`.
```go
type Visitor interface {
    Enter(node ASTNode, ctx *VisitorContext) bool
    Leave(node ASTNode, ctx *VisitorContext)
}

type VisitorContext struct {
    Compilation      *Compilation
    CurrentScope     *Scope
    CurrentNamespace string
    CurrentFile      *SourceFile
    Ancestors        []ASTNode   // Stack of parent nodes
}
```

## 7. The BindingTable
Because Fénix is declarative, creating a duplicated `BoundTree` is a waste of memory. The Binder simply generates a `BindingTable` mapping `ASTNode` identifiers to a `Binding` record.
```go
type BindingTable struct {
    // Maps a NodeID (e.g., an IdentExpr representing "Customer") 
    // to its resolved semantic binding.
    Bindings map[NodeID]*Binding
}
// Architectural Rule: Once constructed by the Binder, the BindingTable is immutable. 
// Semantic analyzers, code fixes, and generators may only read bindings.

type Binding struct {
    Symbol        Symbol         // Resolved semantic identity
    Type          Symbol         // e.g., TypeSymbol (inferred or explicit)
    ConstantValue ConstantValue  // For constant folding
    Conversion    *Conversion    // Implicit/explicit type conversions
    IsLValue      bool           // Can this binding be assigned to?
    IsImplicit    bool           // Was this binding generated by the compiler?
    Diagnostics   []Diagnostic   // Errors specific to this binding
}

type ConstantValue interface {
    constant()
} // Implemented by StringConstant, IntConstant, FloatConstant, BoolConstant, NullConstant.

type ConversionKind int

const (
    ConversionNone ConversionKind = iota
    ConversionImplicit
    ConversionExplicit
)

type Conversion struct {
    Kind     ConversionKind
    FromType Symbol
    ToType   Symbol
}
```

## 8. The Symbol System (SPEC-306 Preview)
Symbols are first-class citizens decoupled from syntax. The Binder produces them, and the Semantic Analyzer / IR Builder consume them.
```go
type SymbolKind int

const (
    SymbolEntity SymbolKind = iota
    SymbolService
    SymbolType
    SymbolNamespace
)

// Base Symbol Interface
type Symbol interface {
    Name()          string
    Kind()          SymbolKind
    Documentation() string
    Declarations()  []Declaration // Decoupled from Span, can have multiple sources
}

// Declaration maps a Symbol back to its source location/AST without embedding parser logic in the Symbol.
type Declaration struct {
    NodeID   NodeID
    Location Span
}

// Example derived symbol
type EntitySymbol struct {
    name       string
    attributes []*AttributeSymbol
    fields     []*FieldSymbol
    // Implements Symbol...
}
```

## 9. Symbol Table (Graph-Based Scope Model)
Scopes form a **graph** (due to imports), not a simple parent-chain tree.

### Scope
```go
type Scope struct {
    Kind         ScopeKind
    Name         string
    Parent       *Scope                // Lexical parent
    Children     []*Scope
    Imports      []*Scope              // Graph edges
    Symbols      map[string]Symbol     // Local declarations
}
// Scope resolution algorithm: Local Symbols -> Imported Scopes -> Parent Scope
```

## 10. Diagnostics (Stable, Categorized Error Codes)
```go
type DiagnosticCode string

const (
    LEX1001 DiagnosticCode = "LEX1001"  // Unterminated string
    PAR2004 DiagnosticCode = "PAR2004"  // Expected identifier
    SEM3021 DiagnosticCode = "SEM3021"  // Unresolved import path
    BND4001 DiagnosticCode = "BND4001"  // Unknown symbol reference
)

type Diagnostic struct {
    Code        DiagnosticCode
    Severity    Severity
    Range       Span
    Message     string
    Arguments   []any          // For i18n and formatted messages (e.g., "Unknown symbol '{0}'")
    Suggestion  string
    Related     []RelatedInfo
    Fixes       []SuggestedFix
}
```

## 11. Conclusion
The Common Language Infrastructure provides Fénix DSLs with industrial-grade primitives: Arena-allocated memory, resilient `NodeID` based on `StablePath`, Workspace-level compilation, graph-based scopes, the decoupled `BindingTable`, and the foundational `Symbol` system. This paves the way for the Lexical System (SPEC-302).
