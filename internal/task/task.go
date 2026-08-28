package task

import(
	"time"

	"github.com/google/uuid"
)

// represent where a rask is in  it's lifecycle
type Status string

const ( 
	StatusPending Status = "PENDING"
	StatusRunning Status = "RUNNING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed Status = "FAILED"
	StatusRetrying Status = "RETRYING"
	StatusDLQ Status = "DLG"
)

// define a single unit of work submitted to a  queue
type Task struct {
	ID string
	Type string
	Payload []byte
	Status Status
	Attempts int
	MaxAttempts int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// create a new task instance
func New(taskType string, payload []byte) *Task {
	now := time.Now()
	return &Task {
		ID: uuid.NewString(),
		Type: taskType,
		Payload: payload,
		Status: StatusPending,
		Attempts: 0,
		MaxAttempts: 3,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
// ttransitions defines which status changes are legal
var transitions = map[Status][]Status{
	StatusPending: {StatusRunning},
	StatusRunning: {StatusCompleted,StatusFailed, StatusRetrying},
	StatusRetrying: {StatusRunning, StatusDLQ},
	StatusFailed: {StatusRetrying, StatusDLQ},
	StatusCompleted: {},
	StatusDLQ: {},
}
// canTransition checks weatger moving to "to" is a valid state change
func (t *Task) CanTransition(to Status) bool {
	allowed, ok := transitions[t.Status]
	if !ok {
		return false
	}

	for _, s := range allowed {
		if s == to {
			return true
		}
	}

	return false
}

// setSTatus applies a status change if legal returns false otherwise
func (t *Task) SetStatus(to Status) bool {
	if !t.CanTransition(to) {
		return false
	}
	t.Status = to
	t.UpdatedAt = time.Now()
	return true
}
