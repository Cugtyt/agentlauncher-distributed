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

type LLMResponseEvent struct {
	AgentID      string                 `json:"agent_id"`
	RequestEvent LLMRequestEvent        `json:"request_event"`
	Response     []llminterface.Message `json:"response"`
}

type LLMRuntimeErrorEvent struct {
	AgentID      string          `json:"agent_id"`
	Error        string          `json:"error"`
	RequestEvent LLMRequestEvent `json:"request_event"`
}
