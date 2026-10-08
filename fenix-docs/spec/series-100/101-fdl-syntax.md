# SPEC-101: Fénix Definition Language (FDL)

## 1. Executive Summary
Markdown is designed for human documentation, not deterministic system execution. Passing Markdown to an LLM introduces severe ambiguity, resulting in hallucinations and unpredictable outputs. 

To solve this, the Fénix Operating System introduces the **Fénix Definition Language (FDL)**. 
FDL is the exclusive language used by humans to dictate intent to the Engineering Runtime. It features a YAML-inspired, human-readable syntax, but is backed by a strict formal grammar and a dedicated Go Parser. The Parser generates an **Engineering AST**, ensuring structural validity before any AI inference occurs.

## 2. Design Principles
- **Indentation-Based Scope:** Like Python or YAML, FDL uses indentation to define scope, keeping it lightweight and free of excessive brackets.
- **Strict Keywords:** FDL enforces specific keywords (`product`, `vision`, `feature`, `story`, `requirement`, `architecture`) mapped directly to the Engineering Runtime Processors.
- **Fail-Fast Parsing:** If the user typos a keyword or violates the hierarchy (e.g., nesting a `story` outside a `feature`), the Go Parser instantly throws a Syntax Error. No AI tokens are wasted on invalid structures.

## 3. The Grammar Hierarchy
The language is designed to naturally flow down the SDLC:
`product` -> `features` -> `stories` -> `requirements` & `architecture` & `tests` -> `release`

## 4. Syntax Example

```fdl
product BookingAutomation

vision
    Automate the reservation process for small clinics, eliminating manual phone calls.

users
    Administrator
    Customer

constraints
    OAuth2
    AzureAD
    GDPR
    MaxLatency: 200ms

feature Authentication
    story Login
        requirement
            User authenticates using Azure AD.
            Session expires after 15 minutes of inactivity.
        acceptance
            Given a valid Azure AD token
            When the user attempts to login
            Then the system issues a JWT and redirects to Dashboard
        architecture
            API
                AuthenticationService
                JwtProvider
                UserRepository
        tests
            Integration: Verify AzureAD handshake
            Unit: JWT decoding and validation

feature Scheduling
    story BookAppointment
        requirement
            Customer selects an available time slot.
            Slot must be locked for 5 minutes during checkout to prevent double-booking.
        architecture
            Database
                PostgreSQL (Isolation Level: Serializable)
            Cache
                Redis (Distributed Lock)

implementation
    Backend: Go (Chi router)
    Frontend: React (Next.js)
    Mobile: Flutter

release
    v1.0
```

## 5. The Parser Pipeline (Go)
Fénix implements a Lexer and Parser in Go to transform this text.

### The Pipeline
1. **FDL File:** The user creates `booking.fdl`.
2. **Lexical Analysis (Lexer):** Tokenizes keywords, identifiers, and indentation blocks.
3. **Syntax Analysis (Parser):** Builds the `Engineering AST` (Abstract Syntax Tree) using Go structs.
4. **Validation:** Checks semantic rules (e.g., a `story` must have at least one `requirement`).
5. **AST Materialization:** The AST is wrapped in the `ExecutionPackage` and passed to the `Product Discovery Processor`.

### Go AST Example
```go
package fdl

type ProductAST struct {
    Name        string
    Vision      string
    Users       []string
    Constraints []Constraint
    Features    []FeatureAST
}

type FeatureAST struct {
    Name    string
    Stories []StoryAST
}

type StoryAST struct {
    Name         string
    Requirements []string
    Acceptance   []string
    Architecture []ArchitectureNode
    Tests        []string
}
```

## 6. The Multi-IR Translation
Once the `Engineering AST` is generated and validated, it is fed into the pipeline:
1. **Product AST** is consumed by the **Discovery Processor**, which outputs the `Product IR`.
2. **Product IR** is consumed by the **Requirements Processor**, which outputs the `Requirements IR`.
3. **Requirements IR** is consumed by the **Architecture Processor**, generating the `Architecture IR`.

Because the entry point (FDL) is strictly typed and parsed, the LLMs (Reasoners) have a mathematically clean baseline to reason upon, drastically reducing hallucination rates and increasing deterministic output.

## 7. Conclusion
FDL bridges the gap between human expressiveness and machine determinism. By forcing the user to map their thoughts into FDL's strict grammar, Fénix eliminates the "garbage-in, garbage-out" problem that plagues traditional AI coding assistants. FDL is the foundation upon which the Fénix Intelligence sits.
