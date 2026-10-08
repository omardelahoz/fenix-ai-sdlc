# Fénix AI SDLC: Execution Engine Specification

## 1. Executive Summary & Core Philosophy
The **Execution Engine** is the heart of the Fénix AI SDLC platform. It is the deterministic Go-based runtime that orchestrates the entire software development lifecycle. Fénix is fundamentally **NOT** a collection of prompts nor an AI agent itself. It is a strict, rules-based engine that utilizes AI Models (Providers) purely as "reasoning coprocessors". 

Agents never read files. Agents never speak to the user. Agents never decide the workflow. The Execution Engine handles all state, persistence, validation, context assembly, and parallel scheduling.

## 2. The 15 Sub-Engines Topology
The Execution Engine is composed of 15 tightly integrated sub-engines, forming a micro-kernel architecture:

1. **Session Manager:** Manages the lifecycle of a user interaction (Session ID, Active Project, Chosen AI Provider, Current Pipeline).
2. **Intent Resolver:** Intercepts raw user input (`/fenix Quiero...`) and categorizes it strictly into a system intent (NewFeature, BugFix, Discovery, Clarification).
3. **Context Builder:** The most critical differentiator. Instead of feeding the whole project to the LLM, it builds a surgical `Context Object` containing only the relevant subset of files, memory, and rules required for the current task.
4. **Workflow Engine:** Translates the resolved intent into a Directed Acyclic Graph (DAG) of high-level phases (e.g. Discovery -> Requirements -> Planning).
5. **Execution Planner:** Breaks down a Workflow phase into granular, executable Tasks (e.g., [Generate Vision, Generate Scope, Generate Personas]).
6. **Agent Runtime:** Injects the `Context Object` into an Agent Definition, delegates the execution, and receives the structured Output Contract.
7. **Skill Runtime:** Intercepts `InvokeSkill` events from Agents (e.g., `GenerateController`) and executes deterministic local tools or scripts, returning the result to the Agent.
8. **Provider Runtime:** Abstracts the LLM APIs (Claude, Gemini, OpenAI). Translates Fénix Prompts into provider-specific API calls.
9. **Scheduler:** The parallel execution coordinator. Dispatches independent Tasks concurrently (e.g., Vision Agent and Stakeholder Agent run at the same time).
10. **Validation Engine:** Intercepts Agent outputs and runs deterministic checks (Compile, Linter, Test Coverage). Triggers automatic retries (up to 3 times) before escalating.
11. **Artifact Engine:** Manages the physical creation, updating, and parsing of Markdown/Code files.
12. **Memory Engine:** Translates engineering decisions into retrievable knowledge. Does not store conversations; stores facts (e.g., "Authentication = Azure AD").
13. **State Engine:** Persists the current pipeline state, error logs, and execution pointers so the engine can resume exactly where it left off after a crash.
14. **Clarification Engine (Manager):** Aggregates `waiting_for_input` states from multiple agents, deduplicates questions, presents a single form to the user, and persists responses in `clarifications/answered.md`.
15. **Report Engine:** Formats final outputs and metrics for user consumption.

## 3. The Request Lifecycle
The lifecycle of a single `/fenix` command follows a strict chronological path:

1. **Input Reception:** `Adapter -> Fenix Runtime -> Session Manager`
2. **Intent & State Resolution:** `Project Loader -> State Engine -> Intent Resolver`
3. **Planning Phase:** `Workflow Engine -> Execution Planner`
4. **Dispatch Phase:** `Scheduler` begins processing the DAG.
5. **Contextualization:** For each Task, `Context Builder` + `Prompt Builder` assemble the payload.
6. **Execution Loop:** `Agent Runtime -> Provider Runtime -> AI Response`
7. **Validation Loop:** `Validation Engine -> Retry (if fail)` or `Skill Runtime (if skill invoked)`
8. **Clarification Loop:** If an agent halts for clarification, `Clarification Engine` queues the question.
9. **Commit Phase:** `Artifact Engine` + `Memory Engine` + `State Engine` persist the results.
10. **Resume/Respond:** The Engine moves to the next Workflow step or responds to the user.

## 4. Prompt Builder & Context Builder Mechanics
Agents do **not** have a `prompt.md`. They have a `behavior.md`. The Execution Engine dynamically constructs the prompt for the AI Provider using a strict universal formula:

```
[SYSTEM]
You are Fénix Agent: {Agent.Name}
Mission: {Agent.Mission}

[BEHAVIOR]
{Agent.Behavior}

[RULES]
Forbidden: {Agent.Constraints.Forbidden}
Allowed: {Agent.Constraints.Allowed}

[CONTEXT]
{ContextBuilder.Output} (Surgical injection of Story, ADRs, Architecture)

[SKILLS]
Available Skills: {Agent.Skills}

[OUTPUT_CONTRACT]
You must respond strictly in JSON fulfilling this schema: {Agent.OutputSchema}
```

The **Context Builder** guarantees token efficiency. If a Backend Agent is working on `AUTH-12`, the Context Builder will only inject the Acceptance Criteria for `AUTH-12`, the `Security-ADR`, and the existing `UserController.cs`. It will **not** inject frontend routing code.

## 5. The Clarification Protocol (Questions as Artifacts)
When an agent (e.g., Database Expert) encounters ambiguity (e.g., PostgreSQL vs SQL Server), it does not generate text asking the user. It generates a structured `ClarificationTask`.

The `Scheduler` pauses that specific branch of the DAG. The `Clarification Engine` collects this task, waits for other parallel agents to finish their current cycle, and groups all doubts into a single interactive prompt for the user.
Once answered, the engine writes to `_ai_sdlc/clarifications/answered.md`, updates the `Product Memory`, and re-awakens the paused agents.

## 6. Validation and the Retry Matrix
Fénix enforces quality through deterministic validation.
When an Agent returns source code:
1. `Validation Engine` runs `go build` or `dotnet build`.
2. If it fails, the compiler error is captured.
3. The `Agent Runtime` automatically reinvokes the AI Provider: *"Your previous output failed with this error: [Error]. Fix it."*
4. This repeats up to `MAX_RETRIES` (default: 3). If it still fails, the task is marked as `Failed` in the `State Engine` and escalated to the user.

## 7. Conclusion
The Execution Engine turns the probabilistic nature of LLMs into a deterministic, enterprise-grade software factory. By isolating context, abstracting providers, and enforcing rigorous state machines, Fénix scales to complex, multi-year projects without degrading context quality.
