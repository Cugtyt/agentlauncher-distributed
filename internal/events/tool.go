package events

import "time"

// ToolDefinition represents a tool that can be called
type ToolDefinition struct {
    Type     string `json:"type"`
    Function struct {
        Name        string      `json:"name"`
        Description string      `json:"description"`
        Parameters  interface{} `json:"parameters"`
    } `json:"function"`
}

// ToolCall represents a function call from LLM
type ToolCall struct {
    ID       string `json:"id"`
    Type     string `json:"type"`
    Function struct {
        Name      string `json:"name"`
        Arguments string `json:"arguments"`
    } `json:"function"`
}

// ToolsExecRequestEvent - Request to execute tools
type ToolsExecRequestEvent struct {
    AgentID   string     `json:"agent_id"`
    ToolCalls []ToolCall `json:"tool_calls"`
    Timestamp time.Time  `json:"timestamp"`
}

// ToolExecStartEvent - Single tool execution starts
type ToolExecStartEvent struct {
    AgentID    string   `json:"agent_id"`
    ToolCallID string   `json:"tool_call_id"`
    ToolName   string   `json:"tool_name"`
    Arguments  string   `json:"arguments"`
    Timestamp  time.Time `json:"timestamp"`
}

// ToolExecFinishEvent - Single tool execution completes
type ToolExecFinishEvent struct {
    AgentID    string      `json:"agent_id"`
    ToolCallID string      `json:"tool_call_id"`
    ToolName   string      `json:"tool_name"`
    Result     interface{} `json:"result"`
    Error      string      `json:"error,omitempty"`
    Timestamp  time.Time   `json:"timestamp"`
}

// ToolsExecResultsEvent - All tool executions complete
type ToolsExecResultsEvent struct {
    AgentID   string       `json:"agent_id"`
    Results   []ToolResult `json:"results"`
    Timestamp time.Time    `json:"timestamp"`
}

// ToolResult represents the result of a single tool execution
type ToolResult struct {
    ToolCallID string      `json:"tool_call_id"`
    ToolName   string      `json:"tool_name"`
    Result     interface{} `json:"result"`
    Error      string      `json:"error,omitempty"`
}