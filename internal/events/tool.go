package events

type ToolCall struct {
	AgentID    string         `json:"agent_id"`
	ToolName   string         `json:"tool_name"`
	ToolCallID string         `json:"tool_call_id"`
	Arguments  map[string]any `json:"arguments"`
}

type ToolResult struct {
	AgentID    string `json:"agent_id"`
	ToolName   string `json:"tool_name"`
	ToolCallID string `json:"tool_call_id"`
	Result     string `json:"result"`
}

type ToolsExecRequestEvent struct {
	AgentID   string     `json:"agent_id"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

func (e ToolsExecRequestEvent) Subject() string { return ToolExecRequestEventName }

type ToolsExecResultsEvent struct {
	AgentID     string       `json:"agent_id"`
	ToolResults []ToolResult `json:"tool_results"`
}

func (e ToolsExecResultsEvent) Subject() string { return ToolExecResultsEventName }

type ToolRuntimeErrorEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

func (e ToolRuntimeErrorEvent) Subject() string { return ToolRuntimeErrorEventName }

type ToolExecStartEvent struct {
	AgentID    string         `json:"agent_id"`
	ToolCallID string         `json:"tool_call_id"`
	ToolName   string         `json:"tool_name"`
	Arguments  map[string]any `json:"arguments"`
}

func (e ToolExecStartEvent) Subject() string { return ToolExecStartEventName }

type ToolExecFinishEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
	Result     string `json:"result"`
}

func (e ToolExecFinishEvent) Subject() string { return ToolExecFinishEventName }

type ToolExecErrorEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
	Error      string `json:"error"`
}

func (e ToolExecErrorEvent) Subject() string { return ToolExecErrorEventName }
