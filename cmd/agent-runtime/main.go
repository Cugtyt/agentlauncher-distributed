package main

import (
    "context"
    "encoding/json"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
    "github.com/cugtyt/agentlauncher-distributed/internal/events"
    "github.com/cugtyt/agentlauncher-distributed/internal/handlers"
    "github.com/cugtyt/agentlauncher-distributed/internal/store"
)

type AgentRuntime struct {
    eventBus     *eventbus.DistributedEventBus
    agentStore   *store.AgentStore
    messageStore *store.MessageStore
    handler      *handlers.AgentHandler
}

func NewAgentRuntime(eventBus *eventbus.DistributedEventBus, agentStore *store.AgentStore, messageStore *store.MessageStore) *AgentRuntime {
    handler := handlers.NewAgentHandler(eventBus, agentStore, messageStore)
    
    return &AgentRuntime{
        eventBus:     eventBus,
        agentStore:   agentStore,
        messageStore: messageStore,
        handler:      handler,
    }
}

func (ar *AgentRuntime) Start() error {
    // Subscribe to task creation events
    err := ar.eventBus.Subscribe("task.create", "agent-runtime", func(data []byte) {
        var event events.TaskCreateEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal task create event: %v", err)
            return
        }
        ar.handler.HandleTaskCreate(context.Background(), event)
    })
    if err != nil {
        return err
    }

    // Subscribe to agent creation events
    err = ar.eventBus.Subscribe("agent.create", "agent-runtime", func(data []byte) {
        var event events.AgentCreateEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal agent create event: %v", err)
            return
        }
        ar.handler.HandleAgentCreate(context.Background(), event)
    })
    if err != nil {
        return err
    }

    // Subscribe to agent start events
    err = ar.eventBus.Subscribe("agent.start", "agent-runtime", func(data []byte) {
        var event events.AgentStartEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal agent start event: %v", err)
            return
        }
        ar.handler.HandleAgentStart(context.Background(), event)
    })
    if err != nil {
        return err
    }

    // Subscribe to LLM response events
    err = ar.eventBus.Subscribe("llm.response", "agent-runtime", func(data []byte) {
        var event events.LLMResponseEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal LLM response event: %v", err)
            return
        }
        ar.handler.HandleLLMResponse(context.Background(), event)
    })
    if err != nil {
        return err
    }

    // Subscribe to tool execution results
    err = ar.eventBus.Subscribe("tool.result", "agent-runtime", func(data []byte) {
        var event events.ToolsExecResultsEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal tool result event: %v", err)
            return
        }
        ar.handler.HandleToolResult(context.Background(), event)
    })

    return err
}

func main() {
    natsURL := getEnv("NATS_URL", "nats://localhost:4222")
    redisURL := getEnv("REDIS_URL", "redis://localhost:6379")

    // Initialize event bus
    eventBus, err := eventbus.NewDistributedEventBus(natsURL)
    if err != nil {
        log.Fatalf("Failed to initialize event bus: %v", err)
    }

    // Initialize stores
    agentStore := store.NewAgentStore(redisURL)
    messageStore := store.NewMessageStore(redisURL)

    // Create agent runtime
    runtime := NewAgentRuntime(eventBus, agentStore, messageStore)

    // Start subscriptions
    if err := runtime.Start(); err != nil {
        log.Fatalf("Failed to start agent runtime: %v", err)
    }

    log.Println("Agent Runtime started successfully")

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down Agent Runtime...")

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Close connections
    eventBus.Close()
    agentStore.Close()
    messageStore.Close()

    log.Println("Agent Runtime stopped")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}