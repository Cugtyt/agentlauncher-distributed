package events

import "time"

// TaskCreateEvent - Initial event when user creates a task
type TaskCreateEvent struct {
    AgentID   string            `json:"agent_id"`
    Task      string            `json:"task"`
    Context   string            `json:"context,omitempty"`
    Tools     []string          `json:"tools,omitempty"`
    MaxSteps  int               `json:"max_steps,omitempty"`
    Metadata  map[string]string `json:"metadata,omitempty"`
    Timestamp time.Time         `json:"timestamp"`
}

// TaskFinishEvent - When task is completed
type TaskFinishEvent struct {
    AgentID   string            `json:"agent_id"`
    Status    string            `json:"status"` // "completed", "failed", "timeout"
    Result    interface{}       `json:"result,omitempty"`
    Error     string            `json:"error,omitempty"`
    Metadata  map[string]string `json:"metadata,omitempty"`
    Timestamp time.Time         `json:"timestamp"`
}