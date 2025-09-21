package events

import "time"

// Message represents a single conversation message
type Message struct {
    Role      string            `json:"role"`      // "system", "user", "assistant", "tool"
    Content   string            `json:"content"`
    ToolCalls []ToolCall        `json:"tool_calls,omitempty"`
    ToolCallID string           `json:"tool_call_id,omitempty"`
    Metadata  map[string]string `json:"metadata,omitempty"`
}

// MessagesAddEvent - Add messages to an agent's conversation
type MessagesAddEvent struct {
    AgentID   string    `json:"agent_id"`
    Messages  []Message `json:"messages"`
    Timestamp time.Time `json:"timestamp"`
}

// MessageGetRequestEvent - Request to retrieve messages
type MessageGetRequestEvent struct {
    AgentID   string `json:"agent_id"`
    ReplyTo   string `json:"reply_to"` // Subject to send response to
    RequestID string `json:"request_id"`
}

// MessageGetResponseEvent - Response with messages
type MessageGetResponseEvent struct {
    AgentID   string    `json:"agent_id"`
    Messages  []Message `json:"messages"`
    RequestID string    `json:"request_id"`
}