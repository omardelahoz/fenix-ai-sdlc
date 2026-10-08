package contextbuilder

// ContextBuilder is responsible for assembling the precise, surgical context
// required by an agent for a specific task.
type ContextBuilder interface {
	Build(agentID string, taskID string) (string, error)
}

