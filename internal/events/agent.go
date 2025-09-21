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

type AgentStartEvent struct {
	AgentID string `json:"agent_id"`
}

type AgentFinishEvent struct {
	AgentID string `json:"agent_id"`
	Result  string `json:"result"`
}

type AgentErrorEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

type AgentDeletedEvent struct {
	AgentID string `json:"agent_id"`
}
