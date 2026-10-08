# ADR 0007: Clarification Manager and Questions as Artifacts

## Status
Accepted

## Context
When multiple agents operate in parallel and encounter blockers due to lack of definition (e.g., Should we use JWT or OAuth? PostgreSQL or SQL Server?), asking the user immediately interrupts them with multiple pop-ups and breaks the SDLC flow. 
Additionally, chat responses are lost in ephemeral history, making it difficult to understand why decisions were made weeks later.

## Decision
1. **Agents never interact directly with the user.** When an agent requires clarification, it generates a structured output with a `waiting_for_input` state and emits a `ClarificationTask`.
2. The **Clarification Manager** groups these tasks, eliminates duplicates, detects contradictions, and categorizes questions by priority (Critical, High, Recommended, Optional).
3. The Fénix Core presents the consolidated questions (in an option-based format) to the user a SINGLE time.
4. **Questions as Artifacts:** All questions and answers are persisted to disk (e.g., `_ai_sdlc/clarifications/pending.md` and `answered.md`). When resolved, the Core updates the `Product Memory` and resumes the paused workflows.

## Consequences
- **Positive:** The user experiences a continuous, orderly flow without being bombarded by questions. The reverse engineering of "why things were done" is physically documented in the repository.
- **Negative:** Forces the design of a structured schema (like JSON Schemas or Forms) for LLM responses, ensuring they return doubts programmatically rather than as free-form text.
