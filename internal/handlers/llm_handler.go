package handlers

import (
    "context"
    "log"
    "time"

    "github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
    "github.com/cugtyt/agentlauncher-distributed/internal/events"
    "github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
    "github.com/cugtyt/agentlauncher-distributed/internal/store"
)

type LLMHandler struct {
    eventBus     *eventbus.DistributedEventBus
    messageStore *store.MessageStore
    llmClient    *llminterface.OpenAIClient
}

func NewLLMHandler(eb *eventbus.DistributedEventBus, ms *store.MessageStore, llm *llminterface.OpenAIClient) *LLMHandler {
    return &LLMHandler{
        eventBus:     eb,
        messageStore: ms,
        llmClient:    llm,
    }
}

func (lh *LLMHandler) HandleLLMRequest(ctx context.Context, event events.LLMRequestEvent) {
    log.Printf("[%s] Processing LLM request", event.AgentID)

    // If messages not provided, fetch from store
    messages := event.Messages
    if len(messages) == 0 {
        var err error
        messages, err = lh.messageStore.GetMessages(event.AgentID)
        if err != nil {
            log.Printf("[%s] Failed to get messages: %v", event.AgentID, err)
            lh.sendError(event.AgentID, err)
            return
        }
    }

    // Convert to LLM format
    llmMessages := lh.convertToLLMMessages(messages)

    // Prepare request
    request := llminterface.ChatRequest{
        Model:       event.Model,
        Messages:    llmMessages,
        Temperature: event.Temperature,
        MaxTokens:   event.MaxTokens,
    }

    // Add tools if provided
    if len(event.Tools) > 0 {
        request.Tools = lh.convertToLLMTools(event.Tools)
    }

    // Default model if not specified
    if request.Model == "" {
        request.Model = "gpt-4"
    }

    // Call LLM
    response, err := lh.llmClient.CreateChatCompletion(ctx, request)
    if err != nil {
        log.Printf("[%s] LLM call failed: %v", event.AgentID, err)
        lh.sendError(event.AgentID, err)
        return
    }

    // Convert response to event message
    responseMessage := lh.convertFromLLMMessage(response.Choices[0].Message)

    // Send response event
    responseEvent := events.LLMResponseEvent{
        AgentID:  event.AgentID,
        Response: responseMessage,
        Usage: events.LLMUsage{
            PromptTokens:     response.Usage.PromptTokens,
            CompletionTokens: response.Usage.CompletionTokens,
            TotalTokens:      response.Usage.TotalTokens,
        },
        Timestamp: time.Now(),
    }

    if err := lh.eventBus.Emit("llm.response", responseEvent); err != nil {
        log.Printf("[%s] Failed to emit LLM response: %v", event.AgentID, err)
    }
}

func (lh *LLMHandler) convertToLLMMessages(messages []events.Message) []llminterface.Message {
    llmMessages := make([]llminterface.Message, 0, len(messages))
    
    for _, msg := range messages {
        llmMsg := llminterface.Message{
            Role:    msg.Role,
            Content: msg.Content,
        }

        // Handle tool calls
        if len(msg.ToolCalls) > 0 {
            llmMsg.ToolCalls = make([]llminterface.ToolCall, len(msg.ToolCalls))
            for i, tc := range msg.ToolCalls {
                llmMsg.ToolCalls[i] = llminterface.ToolCall{
                    ID:   tc.ID,
                    Type: tc.Type,
                    Function: llminterface.FunctionCall{
                        Name:      tc.Function.Name,
                        Arguments: tc.Function.Arguments,
                    },
                }
            }
        }

        // Handle tool response
        if msg.ToolCallID != "" {
            llmMsg.ToolCallID = msg.ToolCallID
        }

        llmMessages = append(llmMessages, llmMsg)
    }
    
    return llmMessages
}

func (lh *LLMHandler) convertFromLLMMessage(llmMsg llminterface.Message) events.Message {
    msg := events.Message{
        Role:     llmMsg.Role,
        Content:  llmMsg.Content,
        Metadata: make(map[string]string),
    }

    // Convert tool calls
    if len(llmMsg.ToolCalls) > 0 {
        msg.ToolCalls = make([]events.ToolCall, len(llmMsg.ToolCalls))
        for i, tc := range llmMsg.ToolCalls {
            msg.ToolCalls[i] = events.ToolCall{
                ID:   tc.ID,
                Type: tc.Type,
                Function: struct {
                    Name      string `json:"name"`
                    Arguments string `json:"arguments"`
                }{
                    Name:      tc.Function.Name,
                    Arguments: tc.Function.Arguments,
                },
            }
        }
    }

    return msg
}

func (lh *LLMHandler) convertToLLMTools(tools []events.ToolDefinition) []llminterface.Tool {
    llmTools := make([]llminterface.Tool, 0, len(tools))
    
    for _, tool := range tools {
        llmTool := llminterface.Tool{
            Type: tool.Type,
            Function: llminterface.Function{
                Name:        tool.Function.Name,
                Description: tool.Function.Description,
                Parameters:  tool.Function.Parameters,
            },
        }
        llmTools = append(llmTools, llmTool)
    }
    
    return llmTools
}

func (lh *LLMHandler) sendError(agentID string, err error) {
    responseEvent := events.LLMResponseEvent{
        AgentID:   agentID,
        Error:     err.Error(),
        Timestamp: time.Now(),
    }

    if err := lh.eventBus.Emit("llm.response", responseEvent); err != nil {
        log.Printf("[%s] Failed to emit error response: %v", agentID, err)
    }
}