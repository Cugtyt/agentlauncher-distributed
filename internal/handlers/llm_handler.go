package handlers

import (
	"context"
	"log"

	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
	"github.com/cugtyt/agentlauncher-distributed/internal/events"
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
)

type LLMHandler struct {
	eventBus     *eventbus.DistributedEventBus
	llmProcessor llminterface.LLMProcessor
}

func NewLLMHandler(eb *eventbus.DistributedEventBus, processor llminterface.LLMProcessor) *LLMHandler {
	if processor == nil {
		panic("LLM processor cannot be nil")
	}
	return &LLMHandler{
		eventBus:     eb,
		llmProcessor: processor,
	}
}

func (lh *LLMHandler) HandleLLMRequest(ctx context.Context, event events.LLMRequestEvent) {
	log.Printf("[%s] Processing LLM request", event.AgentID)

	response, err := lh.llmProcessor(event.Messages, event.ToolSchemas, event.AgentID, lh.eventBus)
	if err != nil {
		errorEvent := events.LLMRuntimeErrorEvent{
			AgentID:      event.AgentID,
			Error:        err.Error(),
			RequestEvent: event,
		}
		if emitErr := lh.eventBus.Emit(errorEvent); emitErr != nil {
			log.Printf("[%s] Failed to emit error event: %v", event.AgentID, emitErr)
		}
		return
	}

	responseEvent := events.LLMResponseEvent{
		AgentID:      event.AgentID,
		RequestEvent: event,
		Response:     response,
	}

	if err := lh.eventBus.Emit(responseEvent); err != nil {
		log.Printf("[%s] Failed to emit LLM response: %v", event.AgentID, err)
	}
}

func (lh *LLMHandler) HandleLLMRuntimeError(ctx context.Context, event events.LLMRuntimeErrorEvent) {
	log.Printf("[%s] Handling LLM runtime error: %s", event.AgentID, event.Error)

	if event.RequestEvent.RetryCount < 5 {
		retryEvent := events.LLMRequestEvent{
			AgentID:     event.RequestEvent.AgentID,
			Messages:    event.RequestEvent.Messages,
			ToolSchemas: event.RequestEvent.ToolSchemas,
			RetryCount:  event.RequestEvent.RetryCount + 1,
		}
		if err := lh.eventBus.Emit(retryEvent); err != nil {
			log.Printf("[%s] Failed to emit retry request: %v", event.AgentID, err)
		}
	} else {
		errorResponse := llminterface.ResponseMessageList{
			llminterface.AssistantMessage{Content: "Runtime error: " + event.Error},
		}
		responseEvent := events.LLMResponseEvent{
			AgentID:      event.RequestEvent.AgentID,
			RequestEvent: event.RequestEvent,
			Response:     errorResponse,
		}
		if err := lh.eventBus.Emit(responseEvent); err != nil {
			log.Printf("[%s] Failed to emit error response: %v", event.AgentID, err)
		}
	}
}
