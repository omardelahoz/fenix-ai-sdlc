---
validators:
  before:
    - ArchitectureValidator
  during:
    - SqlSyntaxValidator
  after:
    - MigrationSafetyValidator
    - DataStandardsValidator
---
