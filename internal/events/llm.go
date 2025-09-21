package events

import "time"

// LLMRequestEvent - Request LLM processing
type LLMRequestEvent struct {
    AgentID     string            `json:"agent_id"`
    Messages    []Message         `json:"messages,omitempty"` // Optional, can fetch from store
    Model       string            `json:"model,omitempty"`
    Temperature float32           `json:"temperature,omitempty"`
    MaxTokens   int               `json:"max_tokens,omitempty"`
    Tools       []ToolDefinition  `json:"tools,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
    Timestamp   time.Time         `json:"timestamp"`
}

// LLMResponseEvent - LLM processing result
type LLMResponseEvent struct {
    AgentID    string    `json:"agent_id"`
    Response   Message   `json:"response"`
    Usage      LLMUsage  `json:"usage,omitempty"`
    Error      string    `json:"error,omitempty"`
    Timestamp  time.Time `json:"timestamp"`
}

// LLMUsage represents token usage statistics
type LLMUsage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}