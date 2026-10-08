# ADR 0006: Context Builder Pattern

## Status
Accepted

## Context
Historically, AI development platforms inject the entire workspace or a massive amount of files into the LLM prompt. This rapidly degrades response quality, exhausts token limits, and causes "hallucinations" by distracting the model with irrelevant context.

## Decision
The **Context Builder** pattern is implemented as a core component of the Fénix Runtime.

Instead of an agent (e.g., Backend Developer) receiving indiscriminate access to the project, the Context Builder acts as a surgical filter:
1. The agent defines in its `behavior.md` and configuration the strict dependencies it needs to read (e.g., `Story`, `Architecture`, `ProductMemory`).
2. When the Scheduler queues the task for the agent, it invokes the Context Builder.
3. The Context Builder reads the corresponding files from disk, extracts only the relevant sections, and constructs a "Final Prompt".
4. This consolidated, concise, and isolated Final Prompt is the only payload sent to the AI Provider.

## Consequences
- **Positive:** Drastic reduction in token costs. Extreme LLM precision (lower margin for hallucination). Prevents UI agents from interfering or making decisions based on DB logic that does not concern them.
- **Negative:** Requires maintaining very fine-grained traceability over where each piece of data lives (memory, state, rules) so the Context Builder knows exactly which files to read and parse.
