package handlers

import (
    "context"
    "log"

    "github.com/yourusername/agentlauncher-distributed/internal/eventbus"
    "github.com/yourusername/agentlauncher-distributed/internal/events"
    "github.com/yourusername/agentlauncher-distributed/internal/store"
)

type MessageHandler struct {
    eventBus     *eventbus.DistributedEventBus
    messageStore *store.MessageStore
}

func NewMessageHandler(eb *eventbus.DistributedEventBus, ms *store.MessageStore) *MessageHandler {
    return &MessageHandler{
        eventBus:     eb,
        messageStore: ms,
    }
}

func (mh *MessageHandler) HandleMessageAdd(ctx context.Context, event events.MessagesAddEvent) {
    log.Printf("[%s] Adding %d messages", event.AgentID, len(event.Messages))

    // Store messages in Redis
    if err := mh.messageStore.AddMessages(event.AgentID, event.Messages); err != nil {
        log.Printf("[%s] Failed to add messages: %v", event.AgentID, err)
        return
    }

    // Log message details for debugging
    for _, msg := range event.Messages {
        log.Printf("[%s] Message added - Role: %s, Length: %d", 
            event.AgentID, msg.Role, len(msg.Content))
        
        if len(msg.ToolCalls) > 0 {
            log.Printf("[%s] Message contains %d tool calls", 
                event.AgentID, len(msg.ToolCalls))
        }
    }
}

func (mh *MessageHandler) HandleMessageGet(ctx context.Context, event events.MessageGetRequestEvent) {
    log.Printf("[%s] Getting messages for request %s", event.AgentID, event.RequestID)

    // Get messages from store
    messages, err := mh.messageStore.GetMessages(event.AgentID)
    if err != nil {
        log.Printf("[%s] Failed to get messages: %v", event.AgentID, err)
        // Send empty response
        messages = []events.Message{}
    }

    // Send response
    response := events.MessageGetResponseEvent{
        AgentID:   event.AgentID,
        Messages:  messages,
        RequestID: event.RequestID,
    }

    // Reply to the specified subject
    if err := mh.eventBus.Emit(event.ReplyTo, response); err != nil {
        log.Printf("[%s] Failed to send message response: %v", event.AgentID, err)
    }
}

func (mh *MessageHandler) HandleMessageDelete(ctx context.Context, agentID string) {
    log.Printf("[%s] Deleting messages", agentID)

    // Delete messages from store
    if err := mh.messageStore.DeleteMessages(agentID); err != nil {
        log.Printf("[%s] Failed to delete messages: %v", agentID, err)
    }
}