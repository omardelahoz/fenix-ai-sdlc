# ADR 0002: Artifact Model

## Status
Accepted

## Context
Fénix relies on unstructured Markdown files and raw source code. Without a universal contract, it is impossible for the Runtime to automatically parse, validate, and load these files into memory efficiently. We need a standardized format for every file managed by Fénix.

## Decision
All Fénix definition files (Agents, Workflows, Pipelines, Templates) must use the `.nd` (Node Definition) extension or a strictly formatted Markdown header. 

Every Artifact must begin with a YAML Frontmatter block containing the following mandatory fields:

```yaml
---
kind: [Workflow | Agent | Skill | Artifact | Policy]
id: [unique_string_identifier]
version: [integer]
owner: [fenix | user | system]
status: [draft | active | deprecated]
---
```

### Schemas
Furthermore, every `kind` has a corresponding JSON Schema located in the Fénix system directory (e.g., `schemas/agent.schema.json`).
When Fénix boots, the Validation Engine parses all `.nd` files and verifies them against their Schema. If a file is malformed, Fénix refuses to load it, preventing downstream execution errors.

## Consequences
- **Positive:** Enables a highly robust Registry loader. Fénix can index the entire system in milliseconds without guessing intent. Predictability in LLM outputs, as we can force LLMs to emit artifacts compliant with the JSON Schema.
- **Negative:** Adds verbosity to writing new agents or workflows manually. It requires strict discipline.
