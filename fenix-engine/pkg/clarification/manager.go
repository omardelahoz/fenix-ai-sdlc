package clarification

type QuestionPriority string

const (
	PriorityCritical QuestionPriority = "Critical"
	PriorityHigh     QuestionPriority = "High"
	PriorityMedium   QuestionPriority = "Medium"
	PriorityLow      QuestionPriority = "Low"
)

type Question struct {
	ID       string
	Text     string
	Options  []string
	Priority QuestionPriority
	Origin   string // Which agent asked this?
}

// ClarificationManager intercepts agent doubts and aggregates them for the user.
type ClarificationManager interface {
	AddQuestion(q Question) error
	GetPendingQuestions() ([]Question, error)
	ResolveQuestion(questionID string, answer string) error
}

