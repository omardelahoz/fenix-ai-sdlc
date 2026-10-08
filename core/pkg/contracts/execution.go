package contracts

import "time"

// ExecutionPackage is the immutable Process Control Block (PCB) of the Fénix Kernel.
// Once created, it never changes. Any state transition generates a new version.
type ExecutionPackage struct {
	Metadata  PackageMetadata
	Task      TaskDescriptor
	Executor  ExecutorDescriptor
	Snapshot  ExecutionSnapshot
	Policies  ExecutionPolicies
	Resources ResourceAllocation
	Contracts Contracts
}

// PackageMetadata contains versioning and tracing information for the ExecutionPackage.
type PackageMetadata struct {
	ID            string
	Version       string // PCB Versioning (e.g., v1, v2) for time-travel and auditing
	SchemaVersion string // Version of this struct schema
	KernelVersion string // Version of the Fénix Kernel that created this PCB
	TraceID       string
	CorrelationID string
}

// TaskDescriptor describes the work to be executed.
type TaskDescriptor struct {
	ID   string
	Type string
}

// ExecutorDescriptor abstracts Agents, Skills, Plugins, or Workflow Steps.
// The Kernel is agnostic to the actor type.
type ExecutorDescriptor struct {
	ID           string
	Type         string // "Agent", "Skill", "Plugin"
	Version      string
	Capabilities []string
}

// ExecutionSnapshot is strictly the frozen state of the repo.
// No errors or state are included here.
type ExecutionSnapshot struct {
	AstNodes         []string
	MemoryReferences []string
}

// ExecutionPolicies define the execution constraints and retry logic.
type ExecutionPolicies struct {
	MaxTokens           int
	MaxFiles            int
	MaxAstDepth         int
	CompressionStrategy string
	RetryPolicy         RetryConfig
}

// RetryConfig defines how the Kernel manages retries.
type RetryConfig struct {
	MaxRetries      int
	BackoffStrategy string
}

// ResourceAllocation tracks budget and deadlines decided by the Execution Planner.
type ResourceAllocation struct {
	Provider        string // e.g., "claude-3.5-sonnet"
	EstimatedTokens int
	ReservedTokens  int
	EstimatedCost   float64
	Priority        int
	Deadline        time.Time
}

// Contracts define the expected output format and validation rules.
type Contracts struct {
	ExpectedFormat  string
	JsonSchema      string
	ValidationRules []string
}

// Validate checks if the ExecutionPackage is valid according to kernel invariants.
func (ep *ExecutionPackage) Validate() error {
	// Check required fields
	if ep.Metadata.ID == "" {
		return &ValidationError{Field: "Metadata.ID", Message: "ID is required"}
	}
	if ep.Task.ID == "" {
		return &ValidationError{Field: "Task.ID", Message: "Task ID is required"}
	}
	if ep.Executor.ID == "" {
		return &ValidationError{Field: "Executor.ID", Message: "Executor ID is required"}
	}

	// Validate resource allocation
	if ep.Resources.Deadline.IsZero() {
		return &ValidationError{Field: "Resources.Deadline", Message: "Deadline is required"}
	}
	if ep.Resources.Provider == "" {
		return &ValidationError{Field: "Resources.Provider", Message: "Provider is required"}
	}

	return nil
}

// NewVersion creates a new immutable version of the ExecutionPackage with updated fields.
// This is the only way to "mutate" an ExecutionPackage - by creating a new version.
func (ep *ExecutionPackage) NewVersion(updates func(*ExecutionPackage)) *ExecutionPackage {
	newPkg := ep.deepCopy()
	updates(newPkg)
	newPkg.Metadata.Version = incrementVersion(ep.Metadata.Version)
	return newPkg
}

func (ep *ExecutionPackage) deepCopy() *ExecutionPackage {
	return &ExecutionPackage{
		Metadata:  ep.Metadata,
		Task:      ep.Task,
		Executor:  ep.Executor,
		Snapshot:  ep.Snapshot,
		Policies:  ep.Policies,
		Resources: ep.Resources,
		Contracts: ep.Contracts,
	}
}

func incrementVersion(version string) string {
	if version == "" {
		return "v1"
	}
	// Simple version increment (v1 -> v2, etc.)
	// In production, this would use semantic versioning
	return "v" + version[1:] + "1"
}

// ValidationError represents a validation error in the ExecutionPackage.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
