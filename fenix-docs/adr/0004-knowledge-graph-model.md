# ADR 0004: Engineering Knowledge Graph Model

## Status
Accepted

## Context
Standard AI coding assistants rely on ephemeral context. If a user changes a requirement in a `.md` file, the assistant does not inherently know that the backend controllers, the Swagger documentation, and the unit tests are now invalid unless explicitly prompted. 
Fénix requires a mechanism to propagate architectural and product changes automatically.

## Decision
Fénix replaces "Product Memory" with the **Engineering Knowledge Graph (EKG)**.

The EKG is a directed graph where every artifact produced by Fénix is a node. 

### Nodes and Edges
- **Nodes:** Vision, Requirement, Epic, Story, Task, Source Code File, Test File, Swagger Spec.
- **Edges (Relationships):** Defined as `Implements`, `DependsOn`, `TestedBy`, `DocumentedBy`.

### Invalidation and Regeneration Propagation
When a node is modified (e.g., the User modifies a Business Rule in the `Product Definition`), the EKG Engine performs a traversal:
1. Marks the `Business Rule` node as `Dirty`.
2. Finds all downstream edges (e.g., `Story-12`, `AuthService.go`, `AuthService_test.go`).
3. Marks all downstream nodes as `Dirty`.
4. The Workflow Engine automatically creates a `Regeneration Task` for every `Dirty` node and queues them in the Scheduler.
5. Agents execute the regeneration tasks (e.g., QA Agent rewrites the tests, Backend Agent rewrites the service).

## Consequences
- **Positive:** True autonomy. Fénix acts like a compiler for the entire software lifecycle. Changing a requirement at the top cascades perfectly to the bottom without manual tracking.
- **Negative:** Enormous architectural complexity. The EKG requires a robust graph database (or in-memory DAG representation) and extremely strict traceability. We must parse ASTs to link source code directly back to specific User Stories.
