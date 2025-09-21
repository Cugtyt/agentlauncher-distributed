package events

import (
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

type MessagesAddEvent struct {
	AgentID  string                   `json:"agent_id"`
	Messages llminterface.MessageList `json:"messages"`
}

type MessageStartStreamingEvent struct {
	AgentID string `json:"agent_id"`
}

type MessageDeltaStreamingEvent struct {
	AgentID string `json:"agent_id"`
	Delta   string `json:"delta"`
}

type MessageDoneStreamingEvent struct {
	AgentID string `json:"agent_id"`
	Message string `json:"message"`
}

type MessageErrorStreamingEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

type ToolCallNameStreamingEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
}

type ToolCallArgumentsStartStreamingEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
}

type ToolCallArgumentsDeltaStreamingEvent struct {
	AgentID        string `json:"agent_id"`
	ToolCallID     string `json:"tool_call_id"`
	ArgumentsDelta string `json:"arguments_delta"`
}

type ToolCallArgumentsDoneStreamingEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	Arguments  string `json:"arguments"`
}

type ToolCallArgumentsErrorStreamingEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	Error      string `json:"error"`
}
