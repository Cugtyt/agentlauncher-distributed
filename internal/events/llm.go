package events

type LLMRequestEvent struct {
	AgentID  string           `json:"agent_id"`
	Messages []Message        `json:"messages,omitempty"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
}

type LLMResponseEvent struct {
	AgentID  string  `json:"agent_id"`
	Response Message `json:"response"`
}

type LLMErrorEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}
