---
validators:
  before:
    - StoryValidator
    - ArchitectureValidator
  during:
    - CompileValidator
    - UnitTestValidator
  after:
    - CodingStandardsValidator
    - SecurityValidator
    - DocumentationValidator
    - BackendReviewValidator
---
