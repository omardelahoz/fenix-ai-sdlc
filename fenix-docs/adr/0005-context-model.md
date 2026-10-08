# ADR 0005: Context Model

## Status
Accepted

## Context
The single most critical factor in the success of an AI-driven system is the quality of the context provided to the LLM. Providing the entire workspace causes hallucinations, context window exhaustion, and excessive API costs. Conversely, providing too little context leads to disconnected and erroneous code.

The **Context Model** defines how Fénix builds, reduces, caches, shares, invalidates, versions, and serializes the exact information an Agent needs.

## Decision
The `Context Builder` is the most important component of Fénix. It does not assemble strings; it assembles a semantic `Context Object` (AST-like) that is later serialized by the `Prompt Runtime`.

### 1. Definition of Context
Context in Fénix is not raw text. It is a structured object containing strictly filtered dependencies retrieved from the `Engineering Knowledge Graph` (EKG).
- **Core Intent:** What the task is trying to achieve (e.g., the Story).
- **Hard Dependencies:** Rules, ADRs, existing code files (AST sliced).
- **State Information:** Previous errors, compilation output.

### 2. Construction (Build)
When a Task is scheduled, it requests specific nodes from the EKG (e.g., `Requires: [AuthService.go, ADR-003]`). The Context Builder queries the Graph, fetches the artifacts, and builds the raw `Context Object`.

### 3. Reduction (Slice/Reduce)
Fénix implements aggressive pruning algorithms:
- **AST Slicing:** If an Agent needs to update a single function in a 2000-line file, the Context Builder parses the file's AST, extracts only the relevant function and its struct definitions, and drops the rest.
- **Summarization:** Long historical ADRs are converted to their "Executive Summary" if they are only tangentially related.

### 4. Caching and Sharing
Context compilation is expensive. 
- Fénix leverages native provider caching (e.g., Anthropic's Prompt Caching) by grouping shared context (like Global System Rules and Architecture Guidelines) at the top of the serialized prompt.
- Cross-agent caching: If `Backend Agent` and `QA Agent` are working on the same Story in a parallel `Task Group`, the Context Builder generates the shared product context once and shares the reference.

### 5. Invalidation and Versioning
A cached context becomes invalid if any underlying node in the EKG changes.
- Every artifact has a hash/version. The Context Object stores a signature of its inputs. If `AuthService.go` is modified by another agent, the Context signature breaks, forcing a rebuild.

### 6. Serialization
The final `Context Object` is passed to the `Prompt Runtime`, which serializes it into the final string payload (JSON, XML tags, or raw text depending on the Provider's optimal format). The Agent never sees the raw files, only the surgically serialized context.

## Consequences
- **Positive:** Agents receive perfectly pruned information. Token consumption drops by 80-90%. Hallucinations approach zero. Cost efficiency maximizes.
- **Negative:** Building an AST parser and semantic reducer for multiple programming languages is one of the most technically demanding tasks in the Fénix architecture.
