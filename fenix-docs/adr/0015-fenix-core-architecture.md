# ADR 0005: Fénix Universal Runtime Architecture

## Status
Accepted

## Context
The architecture of Fénix has evolved from the idea of a simple probabilistic orchestrator to a solid, enterprise-grade **Universal Runtime** written in Go.

The interaction model does not rely on Fénix living as a "prompt" inside a chat tool like Claude or Antigravity. Instead, Fénix is an independent CLI engine (`fenix`) that acts deterministically, orchestrating the context, tasks, and delegating reasoning capabilities to LLMs through AI Provider adapters.

## Decision
The master architecture of Fénix SDLC consists of the following high-level elements:

1. **CLI / Adapter (Entrypoint):** The `/fenix` command (used directly in the terminal or through a chat proxy/extension) interacts with the Go Runtime.
2. **Session Manager:** Controls active memory, the current project, and the state of the interaction.
3. **Intent Resolver:** Classifies the intent of the raw user command.
4. **Context Builder (ADR 0006):** Prepares surgical contexts instead of flooding the LLM's context window.
5. **Workflow Engine & Execution Planner:** Translates intents into a Directed Acyclic Graph (DAG) of tasks (e.g., Discovery -> Requirements -> Planning).
6. **Parallel Scheduler:** Sends independent tasks to the queue for concurrent execution.
7. **Clarification Manager (ADR 0007):** Intercepts agents' doubts to consolidate them and ask the user a single time, saving the answers as artifacts on disk.
8. **AI Provider Manager:** Abstracts AI communication. Agents do not know if they are talking to Claude, Gemini, or OpenAI. They emit tasks, and the Provider returns results.

## Consequences
- **Positive:** Total decoupling from the underlying LLM. Immune to time (if GPT-6 emerges tomorrow, Fénix simply implements a new Provider). Extreme token efficiency thanks to the Context Builder. Massive scalability for huge projects.
- **Negative:** The implementation of the Go Runtime is significantly more complex and requires the design of clean interfaces, concurrency management, and a custom task scheduler.
