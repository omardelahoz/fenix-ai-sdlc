# KERNEL ARCHITECTURE FREEZE POLICY

## Status
**ENFORCED** - Effective v1.0

## Directive
The Fénix Microkernel (Scheduler, Memory Manager, Dispatcher, IPC, VFS, Semantic Index, Drivers) is the foundational infrastructure of the Fénix Operating System. Its architecture has reached stability and maturity.

**Rule:**
"The architecture of the Kernel is completely frozen. No new features, components, or conceptual changes will be accepted in the Kernel without a formally approved Architecture Decision Record (ADR) by the Chief Architect."

## Rationale
The Kernel is infrastructure; it allocates resources and executes instructions deterministically. The intelligence, business logic, and AI integrations belong entirely to the **Engineering Runtime**. By freezing the Kernel, we guarantee a rock-solid, bug-free execution base that will not break when the Engineering Runtime evolves.

## Allowed Modifications
- Bug fixes.
- Performance optimizations (e.g., memory mapping improvements, better NUMA alignment).
- Implementation of the exact contracts defined in `SPEC-001` through `SPEC-004`.

## Forbidden Modifications
- Introducing new System Calls that contain business logic (e.g., `SysGenerateCode`).
- Coupling Kernel Managers directly without using the IPC bus.
- Adding AI Provider logic directly into the Kernel without passing through the AI Driver Interface.
