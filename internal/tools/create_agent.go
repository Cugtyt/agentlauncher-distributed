package tools

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/google/uuid"
    "github.com/yourusername/agentlauncher-distributed/internal/eventbus"
    "github.com/yourusername/agentlauncher-distributed/internal/events"
)

type CreateAgentTool struct {
    eventBus *eventbus.DistributedEventBus
}

func NewCreateAgentTool(eventBus *eventbus.DistributedEventBus) *CreateAgentTool {
    return &CreateAgentTool{
        eventBus: eventBus,
    }
}

func (t *CreateAgentTool) Name() string {
    return "create_agent"
}

func (t *CreateAgentTool) Description() string {
    return "Create a sub-agent to handle a specific task"
}

func (t *CreateAgentTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
    task, ok := args["task"].(string)
    if !ok {
        return nil, fmt.Errorf("task argument is required and must be a string")
    }

    // Optional context
    contextStr, _ := args["context"].(string)
    
    // Optional tools for sub-agent
    var tools []string
    if toolsArg, ok := args["tools"].([]interface{}); ok {
        for _, tool := range toolsArg {
            if toolStr, ok := tool.(string); ok {
                tools = append(tools, toolStr)
            }
        }
    }

    // Get parent agent ID from context or args
    parentID, _ := args["parent_id"].(string)
    if parentID == "" {
        // This will be set by the ToolHandler
        parentID = "unknown"
    }

    // Generate sub-agent ID
    subAgentID := fmt.Sprintf("%s_%s", parentID, uuid.New().String()[:8])
    
    log.Printf("CreateAgentTool: Creating sub-agent %s for task: %s", subAgentID, task)

    // Create initial messages for sub-agent
    systemMessage := events.Message{
        Role:    "system",
        Content: "You are a specialized AI assistant created to handle a specific task.",
    }

    userMessage := events.Message{
        Role:    "user",
        Content: task,
    }

    // If context is provided, add it
    if contextStr != "" {
        systemMessage.Content += fmt.Sprintf("\n\nContext: %s", contextStr)
    }

    // Emit messages add event
    messagesEvent := events.MessagesAddEvent{
        AgentID:   subAgentID,
        Messages:  []events.Message{systemMessage, userMessage},
        Timestamp: time.Now(),
    }
    
    if err := t.eventBus.Emit("message.add", messagesEvent); err != nil {
        return nil, fmt.Errorf("failed to add messages: %w", err)
    }

    // Create the sub-agent
    createEvent := events.AgentCreateEvent{
        AgentID:   subAgentID,
        ParentID:  parentID,
        Task:      task,
        Context:   contextStr,
        Tools:     tools,
        Timestamp: time.Now(),
    }

    if err := t.eventBus.Emit("agent.create", createEvent); err != nil {
        return nil, fmt.Errorf("failed to create sub-agent: %w", err)
    }

    // Return success with sub-agent info
    return map[string]interface{}{
        "agent_id":   subAgentID,
        "parent_id":  parentID,
        "task":       task,
        "status":     "created",
        "message":    fmt.Sprintf("Sub-agent %s created successfully", subAgentID),
        "created_at": time.Now().Format(time.RFC3339),
    }, nil
}