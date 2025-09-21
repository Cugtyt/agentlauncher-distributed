package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
	"github.com/cugtyt/agentlauncher-distributed/internal/events"
	"github.com/cugtyt/agentlauncher-distributed/internal/store"
)

type AgentHandler struct {
	eventBus     *eventbus.DistributedEventBus
	agentStore   *store.AgentStore
	messageStore *store.MessageStore
}

func NewAgentHandler(eb *eventbus.DistributedEventBus, as *store.AgentStore, ms *store.MessageStore) *AgentHandler {
	return &AgentHandler{
		eventBus:     eb,
		agentStore:   as,
		messageStore: ms,
	}
}

func (ah *AgentHandler) HandleTaskCreate(ctx context.Context, event events.TaskCreateEvent) {
	log.Printf("[%s] Handling task create: %s", event.AgentID, event.Task)

	// Create initial system message
	systemMessage := events.Message{
		Role:    "system",
		Content: "You are a helpful AI assistant that can use tools to accomplish tasks.",
	}

	// Create user message with the task
	userMessage := events.Message{
		Role:    "user",
		Content: event.Task,
	}

	// Emit messages add event
	messagesEvent := events.MessagesAddEvent{
		AgentID:   event.AgentID,
		Messages:  []events.Message{systemMessage, userMessage},
		Timestamp: time.Now(),
	}

	if err := ah.eventBus.Emit("message.add", messagesEvent); err != nil {
		log.Printf("[%s] Failed to emit messages event: %v", event.AgentID, err)
		return
	}

	// Create agent
	agentCreateEvent := events.AgentCreateEvent{
		AgentID:   event.AgentID,
		Task:      event.Task,
		Context:   event.Context,
		Metadata:  event.Metadata,
		Timestamp: time.Now(),
	}

	if err := ah.eventBus.Emit("agent.create", agentCreateEvent); err != nil {
		log.Printf("[%s] Failed to emit agent create event: %v", event.AgentID, err)
		return
	}
}

func (ah *AgentHandler) HandleAgentCreate(ctx context.Context, event events.AgentCreateEvent) {
	log.Printf("[%s] Creating agent (parent: %s)", event.AgentID, event.ParentID)

	// Store agent in Redis
	if err := ah.agentStore.Create(event.AgentID, event.ParentID); err != nil {
		log.Printf("[%s] Failed to store agent: %v", event.AgentID, err)
		return
	}

	// Update agent with task and context
	updates := map[string]interface{}{
		"task":     event.Task,
		"context":  event.Context,
		"metadata": event.Metadata,
	}

	if err := ah.agentStore.Update(event.AgentID, updates); err != nil {
		log.Printf("[%s] Failed to update agent: %v", event.AgentID, err)
	}

	// Start the agent
	startEvent := events.AgentStartEvent{
		AgentID:   event.AgentID,
		Timestamp: time.Now(),
	}

	if err := ah.eventBus.Emit("agent.start", startEvent); err != nil {
		log.Printf("[%s] Failed to emit agent start event: %v", event.AgentID, err)
		return
	}
}

func (ah *AgentHandler) HandleAgentStart(ctx context.Context, event events.AgentStartEvent) {
	log.Printf("[%s] Agent starting", event.AgentID)

	// Get agent info
	agent, err := ah.agentStore.Get(event.AgentID)
	if err != nil {
		log.Printf("[%s] Failed to get agent info: %v", event.AgentID, err)
		return
	}

	// Get messages for context
	messages, err := ah.messageStore.GetMessages(event.AgentID)
	if err != nil {
		log.Printf("[%s] Failed to get messages: %v", event.AgentID, err)
		// Continue with empty messages
		messages = []events.Message{}
	}

	// Prepare tool definitions if tools are available
	var tools []events.ToolDefinition
	if len(agent.Tools) > 0 {
		tools = ah.prepareToolDefinitions(agent.Tools)
	}

	// Request LLM processing
	llmRequest := events.LLMRequestEvent{
		AgentID:   event.AgentID,
		Messages:  messages,
		Model:     "gpt-4",
		Tools:     tools,
		Timestamp: time.Now(),
	}

	if err := ah.eventBus.Emit("llm.request", llmRequest); err != nil {
		log.Printf("[%s] Failed to emit LLM request: %v", event.AgentID, err)
		ah.handleAgentError(event.AgentID, err)
	}
}

func (ah *AgentHandler) HandleLLMResponse(ctx context.Context, event events.LLMResponseEvent) {
	log.Printf("[%s] Received LLM response", event.AgentID)

	if event.Error != "" {
		log.Printf("[%s] LLM error: %s", event.AgentID, event.Error)
		ah.handleAgentError(event.AgentID, fmt.Errorf("%s", event.Error))
		return
	}

	// Store the assistant's response
	messagesEvent := events.MessagesAddEvent{
		AgentID:   event.AgentID,
		Messages:  []events.Message{event.Response},
		Timestamp: time.Now(),
	}

	if err := ah.eventBus.Emit("message.add", messagesEvent); err != nil {
		log.Printf("[%s] Failed to emit messages event: %v", event.AgentID, err)
	}

	// Check if the response contains tool calls
	if len(event.Response.ToolCalls) > 0 {
		log.Printf("[%s] LLM requested %d tool calls", event.AgentID, len(event.Response.ToolCalls))

		// Execute tools
		toolsRequest := events.ToolsExecRequestEvent{
			AgentID:   event.AgentID,
			ToolCalls: event.Response.ToolCalls,
			Timestamp: time.Now(),
		}

		if err := ah.eventBus.Emit("tool.execute", toolsRequest); err != nil {
			log.Printf("[%s] Failed to emit tools request: %v", event.AgentID, err)
			ah.handleAgentError(event.AgentID, err)
		}
		return
	}

	// No tool calls, agent is complete
	ah.completeAgent(event.AgentID, event.Response.Content)
}

func (ah *AgentHandler) HandleToolResult(ctx context.Context, event events.ToolsExecResultsEvent) {
	log.Printf("[%s] Received tool results: %d results", event.AgentID, len(event.Results))

	// Store tool results as messages
	messages := make([]events.Message, 0, len(event.Results))
	for _, result := range event.Results {
		msg := events.Message{
			Role:       "tool",
			Content:    fmt.Sprintf("%v", result.Result),
			ToolCallID: result.ToolCallID,
			Metadata: map[string]string{
				"tool_name": result.ToolName,
			},
		}
		messages = append(messages, msg)
	}

	// Add tool results to messages
	messagesEvent := events.MessagesAddEvent{
		AgentID:   event.AgentID,
		Messages:  messages,
		Timestamp: time.Now(),
	}

	if err := ah.eventBus.Emit("message.add", messagesEvent); err != nil {
		log.Printf("[%s] Failed to emit messages event: %v", event.AgentID, err)
	}

	// Continue conversation with LLM
	allMessages, err := ah.messageStore.GetMessages(event.AgentID)
	if err != nil {
		log.Printf("[%s] Failed to get messages: %v", event.AgentID, err)
		ah.handleAgentError(event.AgentID, err)
		return
	}

	// Get agent info for tools
	agent, _ := ah.agentStore.Get(event.AgentID)
	var tools []events.ToolDefinition
	if agent != nil && len(agent.Tools) > 0 {
		tools = ah.prepareToolDefinitions(agent.Tools)
	}

	// Request next LLM response
	llmRequest := events.LLMRequestEvent{
		AgentID:   event.AgentID,
		Messages:  allMessages,
		Model:     "gpt-4",
		Tools:     tools,
		Timestamp: time.Now(),
	}

	if err := ah.eventBus.Emit("llm.request", llmRequest); err != nil {
		log.Printf("[%s] Failed to emit LLM request: %v", event.AgentID, err)
		ah.handleAgentError(event.AgentID, err)
	}
}

func (ah *AgentHandler) completeAgent(agentID string, result string) {
	log.Printf("[%s] Agent completing with result", agentID)

	// Update agent status
	updates := map[string]interface{}{
		"status": "completed",
	}
	ah.agentStore.Update(agentID, updates)

	// Get agent info to check if it has a parent
	agent, _ := ah.agentStore.Get(agentID)

	// Emit finish event
	finishEvent := events.AgentFinishEvent{
		AgentID:   agentID,
		Status:    "completed",
		Result:    result,
		Timestamp: time.Now(),
	}

	ah.eventBus.Emit("agent.finish", finishEvent)

	// If this is a sub-agent, don't emit task finish
	if agent != nil && agent.ParentID != "" {
		log.Printf("[%s] Sub-agent completed, parent: %s", agentID, agent.ParentID)
		return
	}

	// This is a top-level agent, emit task finish
	taskFinishEvent := events.TaskFinishEvent{
		AgentID:   agentID,
		Status:    "completed",
		Result:    result,
		Timestamp: time.Now(),
	}

	ah.eventBus.Emit("task.finish", taskFinishEvent)

	// Schedule cleanup
	ah.scheduleCleanup(agentID)
}

func (ah *AgentHandler) handleAgentError(agentID string, err error) {
	log.Printf("[%s] Agent error: %v", agentID, err)

	// Update agent status
	updates := map[string]interface{}{
		"status": "failed",
	}
	ah.agentStore.Update(agentID, updates)

	// Emit finish event with error
	finishEvent := events.AgentFinishEvent{
		AgentID:   agentID,
		Status:    "failed",
		Error:     err.Error(),
		Timestamp: time.Now(),
	}

	ah.eventBus.Emit("agent.finish", finishEvent)

	// Schedule cleanup
	ah.scheduleCleanup(agentID)
}

func (ah *AgentHandler) scheduleCleanup(agentID string) {
	// Schedule deletion after a delay
	go func() {
		time.Sleep(5 * time.Minute)

		log.Printf("[%s] Cleaning up agent", agentID)

		// Delete agent and its messages
		ah.agentStore.Delete(agentID)
		ah.messageStore.DeleteMessages(agentID)

		// Emit deleted event
		deletedEvent := events.AgentDeletedEvent{
			AgentID:   agentID,
			Timestamp: time.Now(),
		}

		ah.eventBus.Emit("agent.deleted", deletedEvent)
	}()
}

func (ah *AgentHandler) prepareToolDefinitions(toolNames []string) []events.ToolDefinition {
	tools := []events.ToolDefinition{}

	for _, toolName := range toolNames {
		switch toolName {
		case "search":
			tools = append(tools, events.ToolDefinition{
				Type: "function",
				Function: struct {
					Name        string      `json:"name"`
					Description string      `json:"description"`
					Parameters  interface{} `json:"parameters"`
				}{
					Name:        "search",
					Description: "Search for information on the web",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"query": map[string]string{
								"type":        "string",
								"description": "The search query",
							},
						},
						"required": []string{"query"},
					},
				},
			})
		case "weather":
			tools = append(tools, events.ToolDefinition{
				Type: "function",
				Function: struct {
					Name        string      `json:"name"`
					Description string      `json:"description"`
					Parameters  interface{} `json:"parameters"`
				}{
					Name:        "weather",
					Description: "Get weather information for a location",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"location": map[string]string{
								"type":        "string",
								"description": "The location to get weather for",
							},
						},
						"required": []string{"location"},
					},
				},
			})
		case "create_agent":
			tools = append(tools, events.ToolDefinition{
				Type: "function",
				Function: struct {
					Name        string      `json:"name"`
					Description string      `json:"description"`
					Parameters  interface{} `json:"parameters"`
				}{
					Name:        "create_agent",
					Description: "Create a sub-agent to handle a specific task",
					Parameters: map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"task": map[string]string{
								"type":        "string",
								"description": "The task for the sub-agent",
							},
							"context": map[string]string{
								"type":        "string",
								"description": "Additional context for the sub-agent",
							},
						},
						"required": []string{"task"},
					},
				},
			})
		}
	}

	return tools
}
