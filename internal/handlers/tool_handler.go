package handlers

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/google/uuid"
    "github.com/yourusername/agentlauncher-distributed/internal/eventbus"
    "github.com/yourusername/agentlauncher-distributed/internal/events"
    "github.com/yourusername/agentlauncher-distributed/internal/tools"
)

type ToolHandler struct {
    eventBus     *eventbus.DistributedEventBus
    toolRegistry *tools.Registry
}

func NewToolHandler(eb *eventbus.DistributedEventBus, tr *tools.Registry) *ToolHandler {
    return &ToolHandler{
        eventBus:     eb,
        toolRegistry: tr,
    }
}

func (th *ToolHandler) HandleToolExecution(ctx context.Context, event events.ToolsExecRequestEvent) {
    log.Printf("[%s] Executing %d tools", event.AgentID, len(event.ToolCalls))

    results := make([]events.ToolResult, 0, len(event.ToolCalls))

    for _, toolCall := range event.ToolCalls {
        // Emit tool start event
        startEvent := events.ToolExecStartEvent{
            AgentID:    event.AgentID,
            ToolCallID: toolCall.ID,
            ToolName:   toolCall.Function.Name,
            Arguments:  toolCall.Function.Arguments,
            Timestamp:  time.Now(),
        }
        
        if err := th.eventBus.Emit("tool.start", startEvent); err != nil {
            log.Printf("[%s] Failed to emit tool start event: %v", event.AgentID, err)
        }

        // Execute the tool
        result := th.executeTool(ctx, event.AgentID, toolCall)
        results = append(results, result)

        // Emit tool finish event
        finishEvent := events.ToolExecFinishEvent{
            AgentID:    event.AgentID,
            ToolCallID: toolCall.ID,
            ToolName:   toolCall.Function.Name,
            Result:     result.Result,
            Error:      result.Error,
            Timestamp:  time.Now(),
        }
        
        if err := th.eventBus.Emit("tool.finish", finishEvent); err != nil {
            log.Printf("[%s] Failed to emit tool finish event: %v", event.AgentID, err)
        }
    }

    // Send all results back
    resultsEvent := events.ToolsExecResultsEvent{
        AgentID:   event.AgentID,
        Results:   results,
        Timestamp: time.Now(),
    }

    if err := th.eventBus.Emit("tool.result", resultsEvent); err != nil {
        log.Printf("[%s] Failed to emit tool results: %v", event.AgentID, err)
    }
}

func (th *ToolHandler) executeTool(ctx context.Context, agentID string, toolCall events.ToolCall) events.ToolResult {
    log.Printf("[%s] Executing tool: %s", agentID, toolCall.Function.Name)

    // Parse arguments
    var args map[string]interface{}
    if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
        return events.ToolResult{
            ToolCallID: toolCall.ID,
            ToolName:   toolCall.Function.Name,
            Error:      fmt.Sprintf("Failed to parse arguments: %v", err),
        }
    }

    // Special handling for create_agent tool
    if toolCall.Function.Name == "create_agent" {
        return th.executeCreateAgent(ctx, agentID, toolCall.ID, args)
    }

    // Get tool from registry
    tool, err := th.toolRegistry.Get(toolCall.Function.Name)
    if err != nil {
        return events.ToolResult{
            ToolCallID: toolCall.ID,
            ToolName:   toolCall.Function.Name,
            Error:      fmt.Sprintf("Tool not found: %v", err),
        }
    }

    // Execute the tool
    result, err := tool.Execute(ctx, args)
    if err != nil {
        return events.ToolResult{
            ToolCallID: toolCall.ID,
            ToolName:   toolCall.Function.Name,
            Error:      fmt.Sprintf("Tool execution failed: %v", err),
        }
    }

    return events.ToolResult{
        ToolCallID: toolCall.ID,
        ToolName:   toolCall.Function.Name,
        Result:     result,
    }
}

func (th *ToolHandler) executeCreateAgent(ctx context.Context, parentAgentID, toolCallID string, args map[string]interface{}) events.ToolResult {
    task, ok := args["task"].(string)
    if !ok {
        return events.ToolResult{
            ToolCallID: toolCallID,
            ToolName:   "create_agent",
            Error:      "Task argument is required",
        }
    }

    context, _ := args["context"].(string)
    
    // Generate sub-agent ID
    subAgentID := fmt.Sprintf("%s_%s", parentAgentID, uuid.New().String()[:8])
    
    log.Printf("[%s] Creating sub-agent: %s", parentAgentID, subAgentID)

    // Create the sub-agent
    createEvent := events.AgentCreateEvent{
        AgentID:   subAgentID,
        ParentID:  parentAgentID,
        Task:      task,
        Context:   context,
        Timestamp: time.Now(),
    }

    if err := th.eventBus.Emit("agent.create", createEvent); err != nil {
        return events.ToolResult{
            ToolCallID: toolCallID,
            ToolName:   "create_agent",
            Error:      fmt.Sprintf("Failed to create sub-agent: %v", err),
        }
    }

    // Wait for sub-agent to complete (simplified - in production use proper async handling)
    // For now, return immediately with agent ID
    return events.ToolResult{
        ToolCallID: toolCallID,
        ToolName:   "create_agent",
        Result: map[string]string{
            "agent_id": subAgentID,
            "status":   "created",
            "message":  fmt.Sprintf("Sub-agent %s created to handle task: %s", subAgentID, task),
        },
    }
}