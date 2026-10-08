# SPEC-204: Versioning & Compatibility

## 1. Executive Summary
Fénix is designed to be a long-living platform. Forcing a monolithic version bump for the entire OS every time a minor plugin API changes is unsustainable.
Fénix implements **Independent ABI & Schema Versioning**. The Kernel, Runtime, Plugin Interfaces, and Artifact Schemas each maintain their own Semantic Versioning (SemVer) lifecycle, allowing decoupled evolution.

## 2. ABI vs Schema
It is critical to distinguish between *Binary Compatibility* (ABI) and *Data Compatibility* (Schema).

### Binary Compatibility (ABI)
Governs the physical interfaces, gRPC payloads, and Go struct signatures.
- **`kernel_abi`**: Governs the struct of `ExecutionPackage` and the EventBus formats. (e.g., `1.0.0`)
- **`runtime_abi`**: Governs the Workflow/Pipeline Orchestration state machines. (e.g., `1.3.0`)
- **`plugin_api`**: Governs the `/proto` gRPC interface and the Plugin Host handshake. (e.g., `2.1.0`)

### Data Compatibility (Schema)
Governs the JSON/YAML structure of files written to disk in the `_fenix/` workspace.
- **`artifact_schema`**: Governs the Canonical Schema (SPEC-110) payload format. (e.g., `1.0.0`)
- **`fdl_schema`**: Governs the specific Fénix Definition Language profile. (e.g., `3.0.0`)

## 3. The Compatibility Matrix (Manifest Declarations)
To ensure system stability, every Plugin (Driver, Capability, or Schema) MUST declare its compatibility matrix in its Universal Manifest (SPEC-109) using standard Go Semantic Versioning constraints (`>=`, `<=`, `^`, `~`).

```yaml
plugin:
  id: "com.fenix.smt_solver"
  version: "1.0.0"
  type: "Capability"

requires:
  plugin_api: ">=2.0.0, <3.0.0"
  runtime_abi: "^1.2.0"
  kernel_abi: ">=1.0.0"
```

When Fénix boots, the `Plugin Host` evaluates this block. If the current `plugin_api` of the core daemon is `1.9.0`, the Plugin is instantly rejected with `ErrIncompatiblePlugin` before any gRPC connection is attempted.

## 4. Build Metadata
In addition to SemVer, the Core Daemon and Plugins must bake in exact build metadata to facilitate debugging, support, and deterministic auditing.

A typical `fenix version` output will look like:
```text
Kernel ABI:   1.2.0
Runtime ABI:  1.3.0
Plugin API:   2.1.0

Build:        4bda81c
Compiler:     Go 1.24
Generated:    2026-07-13T10:14:00Z
```

## 5. Graceful Degradation & Schema Evolution
When the `artifact_schema` or `fdl_schema` bumps to a new major version, existing FDL files in a user's workspace might become technically obsolete.

To handle this, Fénix dictates that **Schema Plugins must ship with Migrators**.
If `Healthcare.fdl` bumps from `v1` to `v2`, the plugin must provide a binary routine to up-convert the AST/IR from `v1` to `v2`. The Runtime detects the version mismatch in the `workspace.json` and automatically executes the migrator, saving the user from manually rewriting thousands of lines of FDL.

## 6. Conclusion
Independent Versioning creates explicit, unbreakable contracts. It allows the core engineering team to rapidly iterate on the `Runtime ABI` without forcing the ecosystem of third-party `Claude` or `SMT` plugin developers to rewrite their code, provided the `plugin_api` constraints are respected.
