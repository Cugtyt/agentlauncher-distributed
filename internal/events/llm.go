package events

import (
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

type LLMRequestEvent struct {
	AgentID     string                    `json:"agent_id"`
	Messages    []llminterface.Message    `json:"messages"`
	ToolSchemas []llminterface.ToolSchema `json:"tool_schemas"`
	RetryCount  int                       `json:"retry_count"`
}

func (e LLMRequestEvent) Subject() string { return LLMRequestEventName }

type LLMResponseEvent struct {
	AgentID      string                 `json:"agent_id"`
	RequestEvent LLMRequestEvent        `json:"request_event"`
	Response     []llminterface.Message `json:"response"`
}

func (e LLMResponseEvent) Subject() string { return LLMResponseEventName }

type LLMRuntimeErrorEvent struct {
	AgentID      string          `json:"agent_id"`
	Error        string          `json:"error"`
	RequestEvent LLMRequestEvent `json:"request_event"`
}

func (e LLMRuntimeErrorEvent) Subject() string { return LLMErrorEventName }
