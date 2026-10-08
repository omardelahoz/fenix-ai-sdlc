# SPEC-100: Engineering Runtime & IR Paradigm

## 1. Executive Summary
Fénix separates the deterministic reality of computing (The Kernel) from the cognitive reality of software development (The Engineering Runtime). 
While the Kernel allocates memory and schedules processes, the **Engineering Runtime** acts as the overarching Intelligence of Fénix. 

Fénix is **not** an "AI Multi-Agent Framework" where humanoid bots chat with each other. It is a highly deterministic pipeline where humans write **FDL (Fénix Definition Language)**, and specialized **Engineering Processors** (powered by AI LLMs acting as compilers/reasoners) transform that knowledge through progressive stages of an **Intermediate Representation (IR)**.

## 2. The Definitive Fénix Architecture

```text
               Fénix Platform
===================================================

       Fénix Definition Language (FDL)
                      │ (Parser)
                      ▼
               Engineering AST
                      │
                      ▼
+-------------------------------------------------+
|              ENGINEERING RUNTIME                |
|                                                 |
|    [Discovery Processor] -> Product IR          |
|    [Planning Processor]  -> Planning IR         |
|    [Architecture Processor] -> Architecture IR  |
|    [Development Processor] -> Code Changes IR   |
|                                                 |
+-------------------------------------------------+
                      │
           Kernel Syscalls (Primitives)
                      │
                      ▼
+-------------------------------------------------+
|              FÉNIX MICROKERNEL                  |
|   (Scheduler, Memory, VFS, Semantic Index)      |
+-------------------------------------------------+
                      │
                  AI Drivers
                      │
                      ▼
            Claude / Gemini / OpenAI
```

## 3. FDL: The End of Markdown Ambiguity
Current AI coding tools rely on Markdown prompts, which are ambiguous and lead to hallucinations. Fénix introduces **FDL (Fénix Definition Language)**.
Inspired by YAML for human readability, FDL possesses a strict formal grammar and a dedicated Parser in Go. 

Humans write FDL. Fénix parses FDL into an **Engineering AST**. This guarantees that before the AI is even invoked, the input is syntactically valid and structure-enforced. FDL is the only language the user speaks to Fénix.

## 4. Engineering Processors: Reasoners, not Agents
In legacy AI systems, an "Agent" is a loose prompt given to an LLM. In Fénix, a runtime plugin registers an **Engineering Processor**.

A Processor's job is not just to compile (A -> B), but to perform a Cognitive Pipeline:
`Input IR -> Reasoning -> Validation -> Transformation -> Output IR`

- **Product Discovery Processor:** `Product AST` -> `Product IR`
- **Architecture Processor:** `Requirements IR` -> `Architecture IR`
- **Development Processor:** `Architecture IR` -> `Development IR`

These processors invoke primitive `Syscalls` (e.g., `ReadArtifact`, `QuerySemantic`) to gather their inputs, keeping all business logic strictly inside the Runtime.

## 5. The Multi-IR Pipeline
Fénix does not have a single IR; it has a sequential pipeline of compatible IRs.
Each step in the Software Development Life Cycle (SDLC) is a transmutation from one IR to a more refined IR.

### The Vendor-Neutral Compilation
1. **PCB Materialization:** The Kernel hands the `ExecutionPackage` (containing the current IR) to the ExecutionManager.
2. **Cognitive Processing:** The target Engineering Processor applies its rules to the IR.
3. **Driver Translation:** The specific AI Driver (e.g., `Claude Driver`) manages the *Session* and translates the generic `Engineering IR` into the provider-specific syntax (XML tags for Claude, JSON schema for OpenAI), executing the inference.

This guarantees absolute independence from the AI Provider. The LLM is treated merely as a backend compilation engine.

## 6. FQL & The Kernel Journal
To support this cognitive runtime, the Kernel provides two massive features:
- **Fénix Query Language (FQL):** Processors query the `Semantic Index` declaratively (`FIND Controllers WHERE Uses(Entity(User))`) rather than reading source code files linearly.
- **The Kernel Journal:** An immutable, `dmesg`-style log. Every Syscall invoked and every state transition is logged. This enables Replay, Audit, Time Travel Debugging, and catastrophic Fault Recovery.

## 7. Conclusion
Fénix is a paradigm shift. It is a platform where FDL dictates the intent, the Engineering AST validates the structure, the Multi-IR Pipeline retains the state, and the Engineering Processors (backed by AI Drivers) perform the cognitive transformations. The Kernel merely facilitates the memory and execution space. Fénix is the first true AI-Native Engineering Operating System.
