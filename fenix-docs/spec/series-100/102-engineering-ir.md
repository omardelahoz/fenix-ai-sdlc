# SPEC-102: Engineering Intermediate Representation (IR) & Type System

## 1. Executive Summary
The most significant point of failure in modern AI coding assistants is relying on raw strings (Markdown, plain text) to represent complex engineering knowledge. 

Fénix eradicates this by introducing the **Engineering Type System** and the **Engineering Intermediate Representation (IR)**. 
Like LLVM, Fénix parses human input (FDL) into an AST, validates it via a Semantic Analyzer, and outputs a strongly typed, object-oriented graph (The Engineering IR). AI Inference Models (Claude, OpenAI) never receive plain text; they receive a serialized, Normalized IR (the equivalent of SSA).

## 2. The Semantic Analyzer Pipeline
A syntax tree (AST) guarantees that the document is well-formed, but not that the *engineering* is valid. Fénix implements a Semantic Analyzer in Go that bridges the gap.

```text
Fénix Definition Language (FDL) 
              │
          (Parser)
              ▼
       Engineering AST (Syntax is valid)
              │
      (Semantic Analyzer) -> e.g., "A Story requires at least 1 Requirement"
              ▼
      Engineering IR (Engineering is valid, Strongly Typed)
```

## 3. The Engineering Type System (Go Interfaces)
Everything in Fénix is a physical, typed object. There are no loose strings. This allows the Kernel's `Semantic Index` and the `Processors` to query and manipulate data deterministically.

### Core Domain Interfaces
```go
package ir

type Node interface {
    ID() string
    Type() NodeType
    Tags() []string
    Hash() string // For cache invalidation
}

type Dependency interface {
    SourceID() string
    TargetID() string
    Relation() RelationType // e.g., "IMPLEMENTS", "TESTS", "DEPENDS_ON"
}
```

### Business Nodes
```go
type RequirementNode struct {
    Node
    Description string
    Priority    PriorityLevel
    Acceptance  []AcceptanceCriteriaNode
}

type StoryNode struct {
    Node
    Title        string
    Requirements []RequirementNode
    Dependencies []Dependency
}
```

### Technical Nodes
```go
type ArchitectureComponentNode struct {
    Node
    ComponentName string
    Layer         ArchitectureLayer // e.g., "API", "Domain", "Infrastructure"
    Exposes       []ApiEndpointNode
}

type ApiEndpointNode struct {
    Node
    Method string
    Path   string
    Inputs []TypeDefinition
}
```

## 4. Multi-IR Evolution
Fénix processes software via sequential IR state changes. The Engineering IR is not static; it evolves as it passes through the Runtime Processors.

1. **Product IR:** Contains `EpicNode`, `UserNode`, `VisionNode`.
2. **Requirements IR:** Contains `StoryNode`, `RequirementNode`, `ConstraintNode`.
3. **Architecture IR:** Contains `ArchitectureComponentNode`, `ApiEndpointNode`, `DatabaseNode`.
4. **Development IR:** Contains `SourceFileNode`, `ClassNode`, `MethodNode`.

*Compatibility Matrix:* Each IR retains pointers (Dependencies) to its parent IR. A `ClassNode` (Dev IR) maintains a hard link to the `RequirementNode` (Req IR) that justified its creation. If the Requirement changes, the Class is marked `Dirty` by the Kernel.

## 5. Normalized Engineering IR (The Fénix SSA)
Before handing the IR to an LLM, Fénix must ensure the LLM doesn't hallucinate or get confused by deep graph traversals. Fénix performs **Normalization** (similar to Static Single Assignment - SSA in LLVM).

The Normalized IR flattens the graph into a highly optimized, deterministic format (usually a strict JSON Schema or XML DOM) specifically tailored for AI context windows.

```text
Engineering IR (Deep Graph, Pointers, Circular Refs)
              │
         (Normalizer)
              ▼
Normalized Engineering IR (Flat, Deduplicated, LLM-Optimized Structure)
              │
       (Provider Driver)
              ▼
   Claude / Gemini / OpenAI
```

By ensuring that Claude and Gemini receive exactly the same Normalized IR schema, Fénix guarantees that the AI acts purely as a reasoning engine, isolated from the complexities of the system's internal graph.

## 6. Conclusion
The Engineering Type System is the absolute contract of Fénix. By enforcing that every concept (Requirement, Architecture, Code) is a strongly typed Go object, Fénix allows `Processors` to execute deterministic logic and validate AI outputs mathematically, preventing the AI from generating structurally invalid code.
