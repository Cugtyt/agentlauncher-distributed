package events

type TaskCreateEvent struct {
	AgentID string `json:"agent_id"`
	Task    string `json:"task"`
}

type TaskFinishEvent struct {
	AgentID string `json:"agent_id"`
	Result  string `json:"result,omitempty"`
}

type TaskErrorEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}
