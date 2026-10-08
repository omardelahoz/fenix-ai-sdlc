# Architecture Freeze Policy

## Status: FROZEN

---

## Series 100 — Engineering Runtime Architecture
**Frozen since:** 2026-07-13

The conceptual design of the Runtime is complete. The boundary between the Microkernel and the Runtime, the orchestration engine logic, the Processor VM lifecycle, the Inference Engine compilation pipeline, and the Plugin gRPC system are final.

### Immutable Specifications
- `SPEC-100` — Engineering IR & Runtime Paradigm
- `SPEC-101` — Fénix Definition Language (FDL)
- `SPEC-102` — Engineering IR Objects
- `SPEC-103` — Orchestration Engines
- `SPEC-104` — Workflow Definition Language (WDL)
- `SPEC-105` — Processor Manifest (PMF)
- `SPEC-106` — Processor Runtime VM
- `SPEC-107` — Inference Engine
- `SPEC-108` — Runtime State Model (Control Plane)
- `SPEC-109` — Runtime Plugin System
- `SPEC-110` — Engineering Artifact Model (Canonical ABI)

---

## Series 200 — Physical Architecture & Implementation Infrastructure
**Frozen since:** 2026-07-13

The physical architecture of the Fénix codebase is complete. The monorepo layout, build system, workspace ABI, core contracts, versioning strategy, dependency injection model, and CLI daemon architecture are final.

### Immutable Specifications
- `SPEC-200` — Repository Layout
- `SPEC-201` — Build System
- `SPEC-202` — Workspace Layout (Workspace ABI)
- `SPEC-203` — Core ABI (Runtime & Kernel Contracts)
- `SPEC-204` — Versioning & Compatibility
- `SPEC-205` — Dependency Injection & Service Composition
- `SPEC-206` — CLI Architecture (Control Plane Interface)

---

## Modification Policy (All Frozen Series)
1. **Allowed:** Bug fixes, typographical corrections, and clarifications.
2. **Forbidden:** Introducing new structural components, changing the capability-first registry model, altering the gRPC plugin boundary, duplicating state between Runtime and Kernel, or restructuring the monorepo layout.
3. **Exceptions:** Any conceptual change or architectural shift MUST be proposed via an Architecture Decision Record (ADR) and requires approval from the Chief Architect.

Signed,
Fénix Chief Architecture Council
