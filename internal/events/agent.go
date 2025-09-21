package events

import "time"

// AgentCreateEvent - When an agent is created (including sub-agents)
type AgentCreateEvent struct {
    AgentID   string            `json:"agent_id"`
    ParentID  string            `json:"parent_id,omitempty"`
    Task      string            `json:"task,omitempty"`
    Context   string            `json:"context,omitempty"`
    Tools     []string          `json:"tools,omitempty"`
    Metadata  map[string]string `json:"metadata,omitempty"`
    Timestamp time.Time         `json:"timestamp"`
}

// AgentStartEvent - When agent begins execution
type AgentStartEvent struct {
    AgentID   string    `json:"agent_id"`
    Timestamp time.Time `json:"timestamp"`
}

// AgentFinishEvent - When agent completes
type AgentFinishEvent struct {
    AgentID   string      `json:"agent_id"`
    Status    string      `json:"status"` // "completed", "failed"
    Result    interface{} `json:"result,omitempty"`
    Error     string      `json:"error,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
}

// AgentDeletedEvent - When agent is cleaned up
type AgentDeletedEvent struct {
    AgentID   string    `json:"agent_id"`
    Timestamp time.Time `json:"timestamp"`
}