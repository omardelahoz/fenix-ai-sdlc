# 🔥 Fénix AI SDLC

> **The AI-Native Engineering Operating System**

Fénix is not just another AI coding framework or agent orchestrator. It is a full-fledged **AI-Native Engineering Operating System** built around declarative specifications rather than imperative scripts. 

Instead of relying on prompt engineering or fragile ad-hoc instructions, products, workflows, and processors in Fénix are described using custom **Domain-Specific Languages (DSLs)**. These languages are compiled into a canonical engineering model, allowing the platform to manage the entire Software Development Life Cycle (SDLC) deterministically, where the orchestration of AI agents is just the final execution step.

---

## 🚀 The Fénix Language System

Fénix introduces three domain-specific languages that form the heart of the system:

- **FDL (Fénix Definition Language):** Describes **WHAT** the product is (Entities, UI, Architecture).
- **WDL (Workflow Definition Language):** Describes **HOW** engineering workflows execute (Pipelines, CI/CD).
- **PMF (Processor Manifest Format):** Describes **HOW** plugins and processors integrate into the platform.

These languages are processed by a custom-built, industrial-grade compiler pipeline that features a Lexer, Parser, AST, Binder, and Semantic Model. 

---

## ⚙️ The Compiler Pipeline

Fénix's biggest innovation is its execution pipeline. Instead of sending raw text to an LLM, Fénix compiles declarative DSLs into structured Intermediate Representations (IR):

```text
FDL / WDL / PMF
       │
       ▼
 Fénix Language System (Lexer, Parser, AST, Binder)
       │
       ▼
 Canonical Language IR
       │
       ▼
 Engineering IR Builder
       │
       ▼
   Fénix Runtime
       │
       ▼
   Execution (AI Agents / Adapters)
```

---

## 🏗 Planned Architecture

```text
                  Fenix AI SDLC

              CLI • IDE • Dashboard

                       │
                       ▼

            Fenix Language System
        (FDL • WDL • PMF Compiler)

                       │
                       ▼

              Canonical Language IR

                       │
                       ▼

              Engineering IR Builder

                       │
                       ▼

                 Fenix Runtime

        ┌────────────┼────────────┐
        │            │            │
 Product Memory  Validation   Report Engine
        │            │            │
        └────────────┼────────────┘
                     │
                     ▼
             Multi-Agent Orchestrator
                     │
                     ▼
          Claude • Codex • Gemini • ...
```

---

## 🧠 Product Memory (Knowledge Graph)

Product Memory in Fénix is not just a document database. It is a rich, structured **Knowledge Graph** containing:

- Specifications & Requirements
- Architecture & Decisions (ADRs)
- The compiled Semantic Model
- Generated Artifacts
- The Engineering IR
- Historical Versions

This completely eliminates context loss, allowing any AI model to seamlessly jump into a project months later and understand the exact state of the system.

---

## 🎯 The Real Differentiator

What makes Fénix fundamentally different from tools like Cursor, Claude Code, Aider, or OpenHands?

- **Custom Compiler:** A bespoke Language System built from scratch.
- **Three DSLs:** FDL, WDL, and PMF to model the entire SDLC.
- **Canonical IR & Engineering IR:** Structured representations instead of raw text prompts.
- **Declarative Architecture:** You declare the end-state, Fénix figures out the execution.
- **Decoupled Runtime:** The Language System is cleanly separated from the Multi-Agent Execution Runtime.
- **Advanced Product Memory:** True semantic understanding of the software product.
- **Model Agnostic:** Swap Claude, Gemini, or local LLMs seamlessly under the Fénix Runtime.

## 🌐 Model Agnostic Architecture

Fénix separates the execution engine from agent definitions, making it possible to support multiple AI ecosystems without modifying the project itself:

### Supported AI Platforms

- **Claude Code** - Anthropic's AI coding assistant
- **Antigravity** - Advanced AI platform
- **Local LLMs** - Run locally hosted models (Ollama, llama.cpp, etc.)
- **Cloud LLM Providers** - OpenAI, Google Gemini, Azure OpenAI, etc.
- **Future AI coding assistants** - Extensible architecture for future platforms

### How It Works

The Fénix Runtime uses adapters to communicate with different AI providers:

```text
Fénix Runtime
    │
    ├─→ Claude Adapter
    ├─→ OpenAI Adapter
    ├─→ Gemini Adapter
    ├─→ Local LLM Adapter
    └─→ Custom Adapter (extensible)
```

This architecture allows you to:
- Switch between AI providers without changing your FDL/WDL files
- Use different models for different tasks (e.g., Claude for architecture, local LLM for code generation)
- Mix and match providers in a single workflow
- Add new AI platforms as they emerge

---

## 📁 Repository Structure

```text
fenix-ai-sdlc/

├── core/          # Core platform utilities
│   ├── language/  # The Fénix Language System (Lexer, Parser, AST, Binder)
│   ├── compiler/  # Semantic Analysis, Canonical IR, Engineering IR
│   ├── runtime/   # Execution Engine, Agent Orchestrator
│   ├── memory/    # Product Memory & Knowledge Graph
│   └── validation/# Quality and Standards Validation
├── adapters/      # Integrations with external AI models (Claude, Gemini, etc.)
├── templates/     # Project and code generation templates
├── definitions/   # Standard library of FDL/WDL/PMF definitions
├── docs/          # Specifications (Series 100, 200, 300) and architecture
└── tools/         # CLI, Dashboard, and internal tooling
```

---

## 📚 Documentation

### Getting Started

- **[QUICKSTART.md](QUICKSTART.md)** - Understand Fénix in 5 minutes
- **[DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md)** - Complete developer guide (setup, architecture, contribution)
- **[ROADMAP.md](ROADMAP.md)** - Implementation roadmap by phases

### Technical Specifications

- **[SPEC-000](fenix-docs/spec/series-000/)** - Kernel API Contracts (ExecutionPackage, Memory Manager)
- **[SPEC-100](fenix-docs/spec/series-100/)** - Engineering Runtime & IR (FDL/WDL/PMF Semantics)
- **[SPEC-200](fenix-docs/spec/series-200/)** - Repository Layout & Build System
- **[SPEC-300](fenix-docs/spec/series-300/)** - Language System Architecture (Lexer, Parser, AST, Binder)
- **[SPEC-400](fenix-docs/spec/series-400/)** - Compiler Host Architecture (Workspace, Snapshots, Scheduler)

### Architecture Decisions

- **[ADRs](fenix-docs/adr/)** - Architecture Decision Records (Operating Model, Domain Model, etc.)

### Examples

- **[FDL Examples](fenix-examples/fdl/)** - Product definition examples (Booking System, Todo App)
- **[WDL Examples](fenix-examples/wdl/)** - Workflow definition examples (SaaS Development, Feature Development)

### Integration

- **[Integration Guide](fenix-docs/INTEGRATION.md)** - How to integrate with Claude, Antigravity, OpenAI, and other AI providers

## 🤝 Contributing

Fénix is currently in the implementation phase following the [ROADMAP.md](ROADMAP.md).

We welcome contributions! Please read the [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for:
- Environment setup
- Code standards
- Testing guidelines
- Contribution workflow

### Current Focus

According to the roadmap, we're prioritizing **Phase 1: Kernel Foundation (SPEC-000)**. See [ROADMAP.md](ROADMAP.md) for details on how to contribute.

---

## 📄 License

License to be defined.
