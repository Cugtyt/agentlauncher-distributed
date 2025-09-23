package tools

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
	"github.com/cugtyt/agentlauncher-distributed/internal/events"
	"github.com/cugtyt/agentlauncher-distributed/internal/handlers"
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
	"github.com/google/uuid"
)

func NewCreateAgentTool(eventBus *eventbus.DistributedEventBus, toolHandler *handlers.ToolHandler) handlers.Tool {
	return handlers.Tool{
		ToolSchema: llminterface.ToolSchema{
			Name:        "create_agent",
			Description: "Create a sub-agent to handle a specific task",
			Parameters: []llminterface.ToolParamSchema{
				{
					Type:        "string",
					Name:        "task",
					Description: "The task for the sub-agent to accomplish",
					Required:    true,
				},
				{
					Type:        "array",
					Name:        "tools",
					Description: "List of tool names that the sub-agent can use",
					Required:    true,
					Items:       map[string]any{"type": "string"},
				},
			},
		},
		Function: func(ctx context.Context, params map[string]any) (string, error) {
			task, ok := params["task"].(string)
			if !ok || task == "" {
				return "", fmt.Errorf("task is required")
			}

			toolsArray, ok := params["tools"].([]string)
			if !ok {
				return "", fmt.Errorf("tools must be an array of strings")
			}

			if len(toolsArray) == 0 {
				return "", fmt.Errorf("tools list cannot be empty")
			}

			for _, toolName := range toolsArray {
				if _, err := toolHandler.GetTool(toolName); err != nil {
					return "", fmt.Errorf("tool '%s' is not available: %w", toolName, err)
				}
			}

			agentID := fmt.Sprintf("subagent-%s", uuid.New().String())
			log.Printf("Creating sub-agent %s with task: %s, tools: %v", agentID, task, toolsArray)

			var toolSchemas []llminterface.ToolSchema
			for _, toolName := range toolsArray {
				if tool, err := toolHandler.GetTool(toolName); err == nil {
					toolSchemas = append(toolSchemas, tool.ToolSchema)
				}
			}

			resultChan := toolHandler.CreateAgentChannel(agentID)

			agentEvent := &events.AgentCreateEvent{
				AgentID:      agentID,
				Task:         task,
				ToolSchemas:  toolSchemas,
				Conversation: []llminterface.Message{},
				SystemPrompt: fmt.Sprintf("You are a sub-agent with the following task: %s", task),
			}

			if err := eventBus.Emit(agentEvent); err != nil {
				toolHandler.RemoveAgentChannel(agentID)
				return "", fmt.Errorf("failed to create sub-agent: %w", err)
			}

			select {
			case result := <-resultChan:
				return result, nil
			case <-ctx.Done():
				toolHandler.RemoveAgentChannel(agentID)
				return "", fmt.Errorf("sub-agent execution cancelled")
			case <-time.After(5 * time.Minute):
				toolHandler.RemoveAgentChannel(agentID)
				return "", fmt.Errorf("sub-agent execution timeout")
			}
		},
	}
}
