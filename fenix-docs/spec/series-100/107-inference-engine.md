# SPEC-107: Inference Engine (The Compiler Pipeline)

## 1. Executive Summary
Most AI platforms treat inference as an HTTP wrapper: `SendPrompt(text) -> Return(String)`. This immediately locks the platform to a specific vendor's prompt formatting (e.g., Anthropic XML vs OpenAI JSON).

In Fénix, the **Inference Engine** is a full compilation subsystem, structurally inspired by LLVM. A Processor never writes a string prompt, and never selects an AI Provider. The Engine receives a strongly typed `Engineering IR`, orchestrates the retrieval of context, optimizes the payload for the token window, selects the optimal AI model, compiles the final prompt, and returns a normalized result.

## 2. The Inference Pipeline
The Inference Engine executes an internal, deterministic DAG for every request:

```text
    [Processor Runtime VM]
             │ (Provides: Normalized IR + Inference Profile)
             ▼
    +-------------------------+
    | 1. Context Builder      | -> Fetches VFS, Journal, Memory
    +-------------------------+
             │
    +-------------------------+
    | 2. Token Budget Manager | -> Compresses, summarizes, limits files
    +-------------------------+
             │
    +-------------------------+
    | 3. Inference Planner    | -> Evaluates Profile (e.g., reasoning=high)
    +-------------------------+
             │
    +-------------------------+
    | 4. Provider Resolver    | -> Selects target Driver (e.g., Claude 3.5)
    +-------------------------+
             │
    +-------------------------+
    | 5. Prompt Compiler      | -> The LLVM analog. Translates IR to XML/JSON
    +-------------------------+
             │
    +-------------------------+
    | 6. Driver Manager       | -> Manages physical Streaming API connection
    +-------------------------+
             │
    +-------------------------+
    | 7. Response Normalizer  | -> Parses output back into Engineering IR
    +-------------------------+
             │
    [Processor Runtime VM]
```

## 3. Context Builder & Token Budget Manager
A Processor does not know how to read files. It simply declares what it needs.
The **Context Builder** uses the `SemanticQuery` service to pull the necessary artifacts from the Kernel VFS. 
Because a context might request 800 files, the **Token Budget Manager** intercepts the payload before it reaches the Prompt Compiler. It applies RAG heuristics, summarizes large files, and hard-crops dependencies to guarantee the LLM's context window is never breached.

## 4. Inference Session & Multi-Step Reasoning
A single `Reason()` call from the Processor Runtime VM may trigger a multi-step loop inside the Inference Engine.
This is encapsulated in an **Inference Session** (analogous to a Database Transaction). 

If the `Response Normalizer` detects that the LLM produced invalid JSON or violated a strict constraint, the Session automatically issues a "Fix and Retry" prompt to the LLM. The Processor VM remains completely oblivious to these internal auto-correction loops.

## 5. Consensus Engine
The Engine supports declarative Ensembles. If a Processor's PMF specifies:
```yaml
inference:
  strategy: majority
  quorum: 3
```
The `Provider Resolver` will route the compiled prompt to GPT-4o, Claude 3.5 Sonnet, and Gemini 1.5 Pro simultaneously. The `Response Normalizer` will evaluate all three and return the consensus result, effectively achieving multi-agent verification without changing a single line of Processor code.

## 6. Streaming-First Architecture & Cache
**All inference is streaming.** Even if the Processor explicitly requests a blocking response, the Driver streams the output to an internal Assembler buffer. This unifies the network architecture.

To drastically reduce costs, the Inference Engine implements an **Inference Cache**. Before invoking the Driver, it hashes the `(Compiled Prompt + Model ID + Temperature)`. If a match is found in the Kernel Cache, the result is returned instantly.

## 7. Conclusion
By modeling Inference as a compilation pipeline rather than an API wrapper, Fénix achieves absolute vendor neutrality. The LLM simply becomes a swappable backend target, akin to x86 or ARM in the LLVM architecture.
