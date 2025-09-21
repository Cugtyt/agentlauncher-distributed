package api

import "time"

// TaskRequest represents a request to create a new task
type TaskRequest struct {
    Task     string            `json:"task"`
    Context  string            `json:"context,omitempty"`
    Tools    []string          `json:"tools,omitempty"`
    MaxSteps int               `json:"max_steps,omitempty"`
    Metadata map[string]string `json:"metadata,omitempty"`
}

// TaskResponse represents the response after creating a task
type TaskResponse struct {
    AgentID   string    `json:"agent_id"`
    Status    string    `json:"status"`
    Message   string    `json:"message"`
    CreatedAt time.Time `json:"created_at"`
}

// AgentStatus represents the current status of an agent
type AgentStatus struct {
    AgentID   string            `json:"agent_id"`
    Status    string            `json:"status"`
    Task      string            `json:"task,omitempty"`
    Progress  int               `json:"progress,omitempty"`
    Result    interface{}       `json:"result,omitempty"`
    Error     string            `json:"error,omitempty"`
    CreatedAt time.Time         `json:"created_at"`
    UpdatedAt time.Time         `json:"updated_at"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

// HealthStatus represents the health of a service
type HealthStatus struct {
    Service   string    `json:"service"`
    Status    string    `json:"status"`
    Version   string    `json:"version"`
    Uptime    string    `json:"uptime"`
    Timestamp time.Time `json:"timestamp"`
}