package events

import (
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

type AgentCreateEvent struct {
	AgentID      string                    `json:"agent_id"`
	Task         string                    `json:"task"`
	ToolSchemas  []llminterface.ToolSchema `json:"tool_schemas"`
	Conversation []llminterface.Message    `json:"conversation"`
	SystemPrompt string                    `json:"system_prompt"`
}

func (e AgentCreateEvent) Subject() string { return AgentCreateEventName }

type AgentStartEvent struct {
	AgentID string `json:"agent_id"`
}

func (e AgentStartEvent) Subject() string { return AgentStartEventName }

type AgentFinishEvent struct {
	AgentID string `json:"agent_id"`
	Result  string `json:"result"`
}

func (e AgentFinishEvent) Subject() string { return AgentFinishEventName }

type AgentErrorEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

func (e AgentErrorEvent) Subject() string { return AgentErrorEventName }

type AgentRuntimeErrorEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

func (e AgentRuntimeErrorEvent) Subject() string { return AgentRuntimeErrorEventName }

type AgentDeletedEvent struct {
	AgentID string `json:"agent_id"`
}

func (e AgentDeletedEvent) Subject() string { return AgentDeletedEventName }
