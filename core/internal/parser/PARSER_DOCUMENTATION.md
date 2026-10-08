# 📝 Parser Module Documentation

## Overview

The Parser module implements the Language System for Fénix (SPEC-300). It provides a unified compiler pipeline for the three DSLs: FDL (Fénix Definition Language), WDL (Workflow Definition Language), and PMF (Processor Manifest Format).

## Architecture

```
core/internal/parser/
├── lexer/              # Lexical analysis
│   ├── token.go        # Token types and definitions
│   └── lexer.go        # Lexer implementation
├── grammar/            # Grammar definitions
│   ├── grammar.go      # Grammar interface
│   ├── fdl_grammar.go  # FDL grammar implementation
│   ├── wdl_grammar.go  # WDL grammar implementation
│   └── pmf_grammar.go  # PMF grammar implementation
├── cst/                # Concrete Syntax Tree (Green/Red Tree) - PENDING
├── ast/                # Abstract Syntax Tree - PENDING
├── binder/             # Binding engine - PENDING
├── semantic/           # Semantic analyzer - PENDING
├── parser.go           # Generic parser engine
└── parser_test.go      # Tests
```

## Compiler Pipeline

```text
Source Text
    │
    ▼
┌──────────────────┐
│  Lexer           │   Tokenizes source into tokens
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Parser Engine   │   Generic LL(1) recursive descent parser
│  + Grammar       │   Grammar provides language-specific rules
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  CST             │   Concrete Syntax Tree (Green/Red Tree) - PENDING
│  (Green/Red)     │   Full-fidelity representation
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  AST             │   Abstract Syntax Tree - PENDING
│  (Canonical)     │   Semantic-only representation
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Binder          │   Resolves references to symbols - PENDING
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Semantic Model  │   Validates semantics - PENDING
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│  Canonical IR    │   Source-agnostic domain model - PENDING
└──────────────────┘
```

## Components

### Lexer (`lexer/`)

The lexer tokenizes source text into tokens with full position tracking.

#### token.go

Defines all token types:

- **Special tokens:** EOF, Error, Comment, Whitespace
- **Literals:** String, Number, Boolean, Null
- **Operators:** =, :, ->, |, +, -, *, /, <, >, <=, >=, ==, !=, &&, ||, !, .
- **Delimiters:** { }, ( ), [ ], <, >, ,, ;
- **FDL keywords:** product, vision, users, constraints, feature, story, requirement, acceptance, architecture, implementation, release
- **WDL keywords:** workflow, on, pipeline, stage, processor, policy, depends, fanout, strategy, gate, timeout, approvers
- **PMF keywords:** processor_manifest, input, output, capability, version, author, description

#### lexer.go

Implements the lexer:

- **NewLexer()** - Creates a new lexer for source text
- **NextToken()** - Returns the next token
- **readChar()** - Reads the next character
- **peek()** - Peeks at the next character
- **readString()** - Reads string literals
- **readNumber()** - Reads number literals
- **readIdentifier()** - Reads identifiers and keywords
- **readComment()** - Reads comment lines
- **skipWhitespace()** - Skips whitespace

**Features:**
- Position tracking (line, column, offset)
- Support for strings (double and single quotes)
- Support for numbers (integers and decimals)
- Support for comments (#)
- Keyword detection

### Grammar (`grammar/`)

The grammar interface defines how different languages are parsed.

#### grammar.go

Defines the Grammar interface:

- **Grammar** - Interface for language-specific parsing rules
- **ParserContext** - Context with lexer and tokens
- **ParseError** - Represents a parsing error
- **Node** - Interface for parse tree nodes
- **NodeType** - Types of nodes

#### fdl_grammar.go

Implements FDL (Fénix Definition Language) grammar:

- **FDLGrammar** - Grammar implementation for FDL
- **ParseRoot()** - Parses FDL document
- **Parsing sections:** vision, users, constraints, features, implementation, release
- **Parsing features and stories**

**Node types:**
- DocumentNode
- VisionNode
- UsersNode
- ConstraintsNode
- FeatureNode
- StoryNode
- ImplementationNode
- ReleaseNode

#### wdl_grammar.go

Implements WDL (Workflow Definition Language) grammar:

- **WDLGrammar** - Grammar implementation for WDL
- **ParseRoot()** - Parses WDL document
- **Parsing triggers:** on events
- **Parsing pipelines:** stages, processors
- **Parsing policies:** retry, resources
- **Parsing fanout:** strategy, processors
- **Parsing gates:** timeout, approvers

**Node types:**
- WorkflowDocumentNode
- TriggerNode
- PipelineNode
- StageNode
- ProcessorNode
- PolicyNode
- RetryPolicyNode
- ResourcesNode
- FanoutNode
- GateNode

#### pmf_grammar.go

Implements PMF (Processor Manifest Format) grammar:

- **PMFGrammar** - Grammar implementation for PMF
- **ParseRoot()** - Parses PMF document
- **Parsing sections:** version, author, description
- **Parsing inputs/outputs:** type, required/optional
- **Parsing capabilities**

**Node types:**
- ProcessorManifestNode
- VersionNode
- AuthorNode
- DescriptionNode
- InputOutputNode
- CapabilityNode

### Parser Engine (`parser.go`)

The generic parser engine that uses grammars:

- **Parser** - Generic LL(1) recursive descent parser
- **NewParser()** - Creates a parser with a grammar
- **Parse()** - Parses the source text
- **advance()** - Advances to the next token
- **expect()** - Expects a specific token type
- **expectKeyword()** - Expects a specific keyword
- **check()** - Checks current token type
- **checkKeyword()** - Checks current keyword
- **match()** - Matches any of multiple token types
- **synchronize()** - Error recovery

**Features:**
- Grammar-agnostic
- Error recovery
- Error reporting with positions
- LL(1) parsing with minimal lookahead

## Usage Examples

### Parsing FDL

```go
import "github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser"

source := `
product BookingSystem

vision
    A modern booking system.

feature AppointmentBooking
    story BookAppointment
`

// Create parser with FDL grammar
fdlGrammar := grammar.NewFDLGrammar()
parser := parser.NewParser(fdlGrammar, source, "example.fdl")

// Parse
doc, err := parser.Parse()
if err != nil {
    log.Fatal(err)
}

// Access parsed document
docNode := doc.(*grammar.DocumentNode)
fmt.Printf("Product: %s\n", docNode.ProductName)
fmt.Printf("Features: %d\n", len(docNode.Features))
```

### Parsing WDL

```go
source := `workflow DevelopmentPipeline
    on ProductUpdated
        pipeline MainPipeline
            stage Discovery
                processor DiscoveryProcessor`

wdlGrammar := grammar.NewWDLGrammar()
parser := parser.NewParser(wdlGrammar, source, "workflow.wdl")

doc, err := parser.Parse()
if err != nil {
    log.Fatal(err)
}

docNode := doc.(*grammar.WorkflowDocumentNode)
fmt.Printf("Workflow: %s\n", docNode.WorkflowName)
fmt.Printf("Pipelines: %d\n", len(docNode.Pipelines))
```

### Parsing PMF

```go
source := `processor_manifest CodeGenerator
    version 1.0
    author John Doe
    capability code_generation`

pmfGrammar := grammar.NewPMFGrammar()
parser := parser.NewParser(pmfGrammar, source, "manifest.pmf")

doc, err := parser.Parse()
if err != nil {
    log.Fatal(err)
}

docNode := doc.(*grammar.ProcessorManifestNode)
fmt.Printf("Processor: %s\n", docNode.Name)
fmt.Printf("Version: %s\n", docNode.Version.Value)
```

### Direct Lexer Usage

```go
import "github.com/omardelahoz/fenix-ai-sdlc/core/internal/parser/lexer"

source := `product BookingSystem`

lex := lexer.NewLexer(source, "example.fdl")
tokens := lex.Tokenize()

for _, token := range tokens {
    fmt.Printf("%s: %s\n", token.Type, token.Value)
}
```

## Testing

```bash
# Run all parser tests
cd core/internal/parser
go test ./...

# Run specific test
go test -run TestParser

# Run with coverage
go test -cover ./...
```

## Current Status

| Component | Status | Progress |
|-----------|--------|----------|
| **Lexer** | ✅ Complete | 100% |
| **Parser Engine** | ✅ Complete | 100% |
| **Grammar Interface** | ✅ Complete | 100% |
| **FDL Grammar** | ✅ Complete | 100% |
| **WDL Grammar** | ✅ Complete | 100% |
| **PMF Grammar** | ✅ Complete | 100% |
| **CST (Green/Red Tree)** | 🔴 Not Started | 0% |
| **AST (Canonical)** | 🔴 Not Started | 0% |
| **Binder** | 🔴 Not Started | 0% |
| **Semantic Analyzer** | 🔴 Not Started | 0% |

## Design Principles

### 1. Unified Parser Engine
The parser engine is generic and uses a Grammar interface. This allows:
- Adding new languages without rewriting parsing logic
- Reusing error recovery and tree building
- Consistent error reporting across languages

### 2. LL(1) Grammar
All grammars are LL(1) with:
- Minimal lookahead (1 token)
- No ambiguity
- Easy to maintain and extend
- Efficient parsing

### 3. Full Fidelity Parsing
The lexer preserves:
- Exact positions (line, column, offset)
- All tokens
- Comments (for future doc-comment preservation)

### 4. Error Recovery
The parser implements:
- Synchronization at block boundaries
- Continuation after errors
- Detailed error messages with positions

## Future Enhancements

### Short Term
- [ ] Implement CST (Green/Red Tree) with structural sharing
- [ ] Implement AST (Canonical) with no parent pointers
- [ ] Implement Binder for symbol resolution
- [ ] Implement Semantic Analyzer for validation

### Long Term
- [ ] Incremental parsing with stable node IDs
- [ ] Arena allocation for GC pressure reduction
- [ ] LSP (Language Server Protocol) integration
- [ ] Syntax highlighting integration
- [ ] Code completion integration

## References

- [SPEC-300](../../fenix-docs/spec/series-300/300-language-system-architecture.md) - Language System Architecture
- [SPEC-302](../../fenix-docs/spec/series-300/302-lexical-system.md) - Lexical System
- [SPEC-303](../../fenix-docs/spec/series-300/303-cst-green-red-tree.md) - CST Green/Red Tree
- [SPEC-304](../../fenix-docs/spec/series-300/304-ast-canonicalization.md) - AST Canonicalization
- [SPEC-305](../../fenix-docs/spec/series-300/305-binding-engine.md) - Binding Engine
- [SPEC-306](../../fenix-docs/spec/series-300/306-symbol-system.md) - Symbol System
