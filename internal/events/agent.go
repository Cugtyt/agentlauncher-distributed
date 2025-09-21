package events

type AgentCreateEvent struct {
	AgentID string   `json:"agent_id"`
	Task    string   `json:"task,omitempty"`
	Context string   `json:"context,omitempty"`
	ToolSet []string `json:"tool_set,omitempty"`
}

type AgentStartEvent struct {
	AgentID string `json:"agent_id"`
}

type AgentFinishEvent struct {
	AgentID string      `json:"agent_id"`
	Result  any `json:"result,omitempty"`
}

type AgentErrorEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

type AgentDeletedEvent struct {
	AgentID string `json:"agent_id"`
}
