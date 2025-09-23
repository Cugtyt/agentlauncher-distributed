package handlers

import (
	"context"
	"log"

	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
	"github.com/cugtyt/agentlauncher-distributed/internal/events"
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
	"github.com/cugtyt/agentlauncher-distributed/internal/store"
)

type AgentHandler struct {
	eventBus              *eventbus.DistributedEventBus
	agentStore            *store.AgentStore
	conversationProcessor func([]llminterface.Message) []llminterface.Message
}

func NewAgentHandler(eb *eventbus.DistributedEventBus, as *store.AgentStore) *AgentHandler {
	return &AgentHandler{
		eventBus:   eb,
		agentStore: as,
	}
}

func (ah *AgentHandler) SetConversationProcessor(processor func([]llminterface.Message) []llminterface.Message) *AgentHandler {
	ah.conversationProcessor = processor
	return ah
}

func (ah *AgentHandler) HandleTaskCreate(ctx context.Context, event events.TaskCreateEvent) {
	agentCreateEvent := events.AgentCreateEvent{
		AgentID:      event.AgentID,
		Task:         event.Task,
		ToolSchemas:  event.ToolSchemas,
		Conversation: event.Conversation,
		SystemPrompt: event.SystemPrompt,
	}

	if err := ah.eventBus.Emit(agentCreateEvent); err != nil {
		log.Printf("[%s] Failed to emit agent create event: %v", event.AgentID, err)
		return
	}
}

func (ah *AgentHandler) HandleAgentCreate(ctx context.Context, event events.AgentCreateEvent) {
	if exists, _ := ah.agentStore.Exists(event.AgentID); exists {
		errorEvent := events.AgentRuntimeErrorEvent{
			AgentID: event.AgentID,
			Error:   "Agent with this ID already exists",
		}
		ah.eventBus.Emit(errorEvent)
		return
	}

	agentData := &store.AgentData{
		AgentID:      event.AgentID,
		Task:         event.Task,
		SystemPrompt: event.SystemPrompt,
		ToolSchemas:  event.ToolSchemas,
		Messages:     event.Conversation,
	}

	if err := ah.agentStore.CreateAgent(agentData); err != nil {
		log.Printf("[%s] Failed to create agent: %v", event.AgentID, err)

		errorEvent := events.AgentErrorEvent{
			AgentID: event.AgentID,
			Error:   err.Error(),
		}
		ah.eventBus.Emit(errorEvent)
		return
	}

	startEvent := events.AgentStartEvent{
		AgentID: event.AgentID,
	}

	if err := ah.eventBus.Emit(startEvent); err != nil {
		log.Printf("[%s] Failed to emit agent start event: %v", event.AgentID, err)
		return
	}
}

func (ah *AgentHandler) HandleAgentStart(ctx context.Context, event events.AgentStartEvent) {
	agent, err := ah.agentStore.GetAgent(event.AgentID)
	if err != nil {
		log.Printf("[%s] Failed to get agent: %v", event.AgentID, err)

		errorEvent := events.AgentErrorEvent{
			AgentID: event.AgentID,
			Error:   err.Error(),
		}
		ah.eventBus.Emit(errorEvent)
		return
	}

	taskMsg := llminterface.NewUserMessage(agent.Task)
	updatedConversation := append(agent.Messages, taskMsg)

	if err := ah.agentStore.SetConversation(event.AgentID, updatedConversation); err != nil {
		log.Printf("[%s] Failed to update conversation: %v", event.AgentID, err)
	}

	messages := llminterface.MessageList{}
	if agent.SystemPrompt != "" {
		systemMsg := llminterface.NewSystemMessage(agent.SystemPrompt)
		messages = append(messages, systemMsg)
	}
	messages = append(messages, updatedConversation...)

	llmRequest := events.LLMRequestEvent{
		AgentID:     event.AgentID,
		Messages:    messages,
		ToolSchemas: agent.ToolSchemas,
		RetryCount:  0,
	}

	if err := ah.eventBus.Emit(llmRequest); err != nil {
		log.Printf("[%s] Failed to emit LLM request: %v", event.AgentID, err)

		errorEvent := events.AgentErrorEvent{
			AgentID: event.AgentID,
			Error:   err.Error(),
		}
		ah.eventBus.Emit(errorEvent)
	}
}

func (ah *AgentHandler) HandleLLMResponse(ctx context.Context, event events.LLMResponseEvent) {
	conversation, err := ah.agentStore.GetConversation(event.AgentID)
	if err != nil {
		log.Printf("[%s] Failed to get conversation: %v", event.AgentID, err)

		errorEvent := events.AgentErrorEvent{
			AgentID: event.AgentID,
			Error:   err.Error(),
		}
		ah.eventBus.Emit(errorEvent)
		return
	}

	updatedConversation := append(conversation, event.Response...)

	if ah.conversationProcessor != nil {
		updatedConversation = ah.conversationProcessor(updatedConversation)
	}

	if err := ah.agentStore.SetConversation(event.AgentID, updatedConversation); err != nil {
		log.Printf("[%s] Failed to update conversation: %v", event.AgentID, err)
	}

	var toolCalls []events.ToolCall
	var finalResponse string

	for _, msg := range event.Response {
		switch msg.Type {
		case llminterface.MessageTypeAssistant:
			finalResponse = msg.Content
		case llminterface.MessageTypeToolCall:
			toolCall := events.ToolCall{
				AgentID:    event.AgentID,
				ToolName:   msg.ToolName,
				ToolCallID: msg.ToolCallID,
				Arguments:  msg.Arguments,
			}
			toolCalls = append(toolCalls, toolCall)
		}
	}

	if len(toolCalls) > 0 {
		toolsRequest := events.ToolsExecRequestEvent{
			AgentID:   event.AgentID,
			ToolCalls: toolCalls,
		}

		if err := ah.eventBus.Emit(toolsRequest); err != nil {
			log.Printf("[%s] Failed to emit tools request: %v", event.AgentID, err)

			errorEvent := events.AgentErrorEvent{
				AgentID: event.AgentID,
				Error:   err.Error(),
			}
			ah.eventBus.Emit(errorEvent)
		}
		return
	}

	finishEvent := events.AgentFinishEvent{
		AgentID: event.AgentID,
		Result:  finalResponse,
	}

	ah.eventBus.Emit(finishEvent)
}

func (ah *AgentHandler) HandleToolResult(ctx context.Context, event events.ToolsExecResultsEvent) {
	agent, err := ah.agentStore.GetAgent(event.AgentID)
	if err != nil {
		log.Printf("[%s] Failed to get agent: %v", event.AgentID, err)

		errorEvent := events.AgentErrorEvent{
			AgentID: event.AgentID,
			Error:   err.Error(),
		}
		ah.eventBus.Emit(errorEvent)
		return
	}

	toolMessages := make(llminterface.MessageList, 0, len(event.ToolResults))
	for _, result := range event.ToolResults {
		msg := llminterface.NewToolResultMessage(result.ToolCallID, result.ToolName, result.Result)
		toolMessages = append(toolMessages, msg)
	}

	updatedConversation := append(agent.Messages, toolMessages...)

	if ah.conversationProcessor != nil {
		updatedConversation = ah.conversationProcessor(updatedConversation)
	}

	if err := ah.agentStore.SetConversation(event.AgentID, updatedConversation); err != nil {
		log.Printf("[%s] Failed to update conversation: %v", event.AgentID, err)
	}

	messages := llminterface.MessageList{}
	if agent.SystemPrompt != "" {
		systemMsg := llminterface.NewSystemMessage(agent.SystemPrompt)
		messages = append(messages, systemMsg)
	}
	messages = append(messages, updatedConversation...)

	llmRequest := events.LLMRequestEvent{
		AgentID:     event.AgentID,
		Messages:    messages,
		ToolSchemas: agent.ToolSchemas,
		RetryCount:  0,
	}

	if err := ah.eventBus.Emit(llmRequest); err != nil {
		log.Printf("[%s] Failed to emit LLM request: %v", event.AgentID, err)

		errorEvent := events.AgentErrorEvent{
			AgentID: event.AgentID,
			Error:   err.Error(),
		}
		ah.eventBus.Emit(errorEvent)
	}
}

func (ah *AgentHandler) HandleAgentFinish(ctx context.Context, event events.AgentFinishEvent) {
	log.Printf("[%s] Agent finished with result: %s", event.AgentID, event.Result)

	deletedEvent := events.AgentDeletedEvent{
		AgentID: event.AgentID,
	}
	ah.eventBus.Emit(deletedEvent)
}

func (ah *AgentHandler) HandleAgentError(ctx context.Context, event events.AgentErrorEvent) {
	log.Printf("[%s] Agent error handled: %s", event.AgentID, event.Error)

	deletedEvent := events.AgentDeletedEvent{
		AgentID: event.AgentID,
	}
	ah.eventBus.Emit(deletedEvent)
}

func (ah *AgentHandler) HandleAgentDeleted(ctx context.Context, event events.AgentDeletedEvent) {
	log.Printf("[%s] Agent deleted", event.AgentID)
	ah.agentStore.Delete(event.AgentID)
}
