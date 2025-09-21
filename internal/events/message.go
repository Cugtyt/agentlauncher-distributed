package events

import (
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

type MessagesAddEvent struct {
	AgentID  string                   `json:"agent_id"`
	Messages llminterface.MessageList `json:"messages"`
}

func (e MessagesAddEvent) Subject() string { return MessageAddEventName }

type MessageStartStreamingEvent struct {
	AgentID string `json:"agent_id"`
}

func (e MessageStartStreamingEvent) Subject() string { return MessageStreamStartEventName }

type MessageDeltaStreamingEvent struct {
	AgentID string `json:"agent_id"`
	Delta   string `json:"delta"`
}

func (e MessageDeltaStreamingEvent) Subject() string { return MessageStreamDeltaEventName }

type MessageDoneStreamingEvent struct {
	AgentID string `json:"agent_id"`
	Message string `json:"message"`
}

func (e MessageDoneStreamingEvent) Subject() string { return MessageStreamDoneEventName }

type MessageErrorStreamingEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

func (e MessageErrorStreamingEvent) Subject() string { return MessageStreamErrorEventName }

type ToolCallNameStreamingEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	ToolName   string `json:"tool_name"`
}

func (e ToolCallNameStreamingEvent) Subject() string { return ToolCallStreamNameEventName }

type ToolCallArgumentsStartStreamingEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
}

func (e ToolCallArgumentsStartStreamingEvent) Subject() string {
	return ToolCallStreamArgsStartEventName
}

type ToolCallArgumentsDeltaStreamingEvent struct {
	AgentID        string `json:"agent_id"`
	ToolCallID     string `json:"tool_call_id"`
	ArgumentsDelta string `json:"arguments_delta"`
}

func (e ToolCallArgumentsDeltaStreamingEvent) Subject() string {
	return ToolCallStreamArgsDeltaEventName
}

type ToolCallArgumentsDoneStreamingEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	Arguments  string `json:"arguments"`
}

func (e ToolCallArgumentsDoneStreamingEvent) Subject() string {
	return ToolCallStreamArgsDoneEventName
}

type ToolCallArgumentsErrorStreamingEvent struct {
	AgentID    string `json:"agent_id"`
	ToolCallID string `json:"tool_call_id"`
	Error      string `json:"error"`
}

func (e ToolCallArgumentsErrorStreamingEvent) Subject() string {
	return ToolCallStreamArgsErrorEventName
}
