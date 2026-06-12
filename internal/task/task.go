package task

import "time"

type Task struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	Priority  int                    `json:"priority"`
	CreatedAt time.Time              `json:"created_at"`
}

func NewTask(id, taskType string, payload map[string]interface{}) Task {
	return Task{
		ID:        id,
		Type:      taskType,
		Payload:   payload,
		Priority:  1,
		CreatedAt: time.Now().UTC(),
	}
}
