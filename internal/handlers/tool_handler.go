package handlers

import (
	"context"
	"fmt"
	"log"

	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
	"github.com/cugtyt/agentlauncher-distributed/internal/events"
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

type Tool struct {
	llminterface.ToolSchema
	Function func(ctx context.Context, args map[string]any) (string, error)
}

type ToolHandler struct {
	eventBus      *eventbus.DistributedEventBus
	tools         map[string]Tool
	agentChannels map[string]chan string
}

func NewToolHandler(eb *eventbus.DistributedEventBus) *ToolHandler {
	return &ToolHandler{
		eventBus:      eb,
		tools:         make(map[string]Tool),
		agentChannels: make(map[string]chan string),
	}
}

func (th *ToolHandler) Register(tool Tool) error {
	if _, exists := th.tools[tool.Name]; exists {
		return fmt.Errorf("tool %s already registered", tool.Name)
	}

	th.tools[tool.Name] = tool
	return nil
}

func (th *ToolHandler) GetTool(name string) (Tool, error) {
	tool, exists := th.tools[name]
	if !exists {
		return Tool{}, fmt.Errorf("tool %s not found", name)
	}

	return tool, nil
}

func (th *ToolHandler) GetAllToolNames() []string {
	names := make([]string, 0, len(th.tools))
	for name := range th.tools {
		names = append(names, name)
	}
	return names
}

func (th *ToolHandler) GetAllToolSchemas() []llminterface.ToolSchema {
	schemas := make([]llminterface.ToolSchema, 0, len(th.tools))
	for _, tool := range th.tools {
		schemas = append(schemas, tool.ToolSchema)
	}
	return schemas
}

func (th *ToolHandler) HandleToolExecution(ctx context.Context, event events.ToolsExecRequestEvent) {
	log.Printf("[%s] Executing %d tools", event.AgentID, len(event.ToolCalls))

	results := make([]events.ToolResult, 0, len(event.ToolCalls))

	for _, toolCall := range event.ToolCalls {
		startEvent := events.ToolExecStartEvent{
			AgentID:    event.AgentID,
			ToolCallID: toolCall.ToolCallID,
			ToolName:   toolCall.ToolName,
			Arguments:  toolCall.Arguments,
		}

		if err := th.eventBus.Emit(startEvent); err != nil {
			log.Printf("[%s] Failed to emit tool start event: %v", event.AgentID, err)
		}

		result := th.executeTool(ctx, event.AgentID, toolCall)
		results = append(results, result)

		finishEvent := events.ToolExecFinishEvent{
			AgentID:    event.AgentID,
			ToolCallID: toolCall.ToolCallID,
			ToolName:   toolCall.ToolName,
			Result:     result.Result,
		}

		if err := th.eventBus.Emit(finishEvent); err != nil {
			log.Printf("[%s] Failed to emit tool finish event: %v", event.AgentID, err)
		}
	}

	resultsEvent := events.ToolsExecResultsEvent{
		AgentID:     event.AgentID,
		ToolResults: results,
	}

	if err := th.eventBus.Emit(resultsEvent); err != nil {
		log.Printf("[%s] Failed to emit tool results: %v", event.AgentID, err)
	}
}

func (th *ToolHandler) executeTool(ctx context.Context, agentID string, toolCall events.ToolCall) events.ToolResult {
	log.Printf("[%s] Executing tool: %s", agentID, toolCall.ToolName)

	var args map[string]any
	if toolCall.Arguments != nil {
		args = toolCall.Arguments
	} else {
		args = make(map[string]any)
	}

	tool, err := th.GetTool(toolCall.ToolName)
	if err != nil {
		errorEvent := events.ToolExecErrorEvent{
			AgentID:    agentID,
			ToolCallID: toolCall.ToolCallID,
			ToolName:   toolCall.ToolName,
			Error:      fmt.Sprintf("Tool not found: %v", err),
		}
		th.eventBus.Emit(errorEvent)

		return events.ToolResult{
			AgentID:    agentID,
			ToolCallID: toolCall.ToolCallID,
			ToolName:   toolCall.ToolName,
			Result:     fmt.Sprintf("Error: Tool not found: %v", err),
		}
	}

	result, err := tool.Function(ctx, args)
	if err != nil {
		errorEvent := events.ToolExecErrorEvent{
			AgentID:    agentID,
			ToolCallID: toolCall.ToolCallID,
			ToolName:   toolCall.ToolName,
			Error:      fmt.Sprintf("Tool execution failed: %v", err),
		}
		th.eventBus.Emit(errorEvent)

		return events.ToolResult{
			AgentID:    agentID,
			ToolCallID: toolCall.ToolCallID,
			ToolName:   toolCall.ToolName,
			Result:     fmt.Sprintf("Error: Tool execution failed: %v", err),
		}
	}

	return events.ToolResult{
		AgentID:    agentID,
		ToolCallID: toolCall.ToolCallID,
		ToolName:   toolCall.ToolName,
		Result:     result,
	}
}

func (th *ToolHandler) CreateAgentChannel(agentID string) chan string {
	resultChan := make(chan string, 1)
	th.agentChannels[agentID] = resultChan
	return resultChan
}

func (th *ToolHandler) RemoveAgentChannel(agentID string) {
	if ch, exists := th.agentChannels[agentID]; exists {
		close(ch)
		delete(th.agentChannels, agentID)
	}
}

func (th *ToolHandler) HandleAgentFinish(agentID string, result string) {
	if ch, exists := th.agentChannels[agentID]; exists {
		select {
		case ch <- result:
		default:
		}
		th.RemoveAgentChannel(agentID)
	}
}
