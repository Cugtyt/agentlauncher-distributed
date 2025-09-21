package events

import (
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

type TaskCreateEvent struct {
	AgentID      string                    `json:"agent_id"`
	Task         string                    `json:"task"`
	ToolSchemas  []llminterface.ToolSchema `json:"tool_schemas"`
	SystemPrompt string                    `json:"system_prompt"`
	Conversation []llminterface.Message    `json:"conversation"`
}

func (e TaskCreateEvent) Subject() string { return TaskCreateEventName }

type TaskFinishEvent struct {
	AgentID string `json:"agent_id"`
	Result  string `json:"result"`
}

func (e TaskFinishEvent) Subject() string { return TaskFinishEventName }

type TaskErrorEvent struct {
	AgentID string `json:"agent_id"`
	Error   string `json:"error"`
}

func (e TaskErrorEvent) Subject() string { return TaskErrorEventName }
