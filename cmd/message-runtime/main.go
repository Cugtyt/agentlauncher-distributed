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

type MessageRuntime struct {
    eventBus     *eventbus.DistributedEventBus
    messageStore *store.MessageStore
    handler      *handlers.MessageHandler
}

func NewMessageRuntime(eventBus *eventbus.DistributedEventBus, messageStore *store.MessageStore) *MessageRuntime {
    handler := handlers.NewMessageHandler(eventBus, messageStore)
    
    return &MessageRuntime{
        eventBus:     eventBus,
        messageStore: messageStore,
        handler:      handler,
    }
}

func (mr *MessageRuntime) Start() error {
    // Subscribe to message add events
    err := mr.eventBus.Subscribe("message.add", "message-runtime", func(ctx context.Context, data []byte) {
        var event events.MessagesAddEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal message add event: %v", err)
            return
        }
        mr.handler.HandleMessageAdd(context.Background(), event)
    })
    if err != nil {
        return err
    }

    // Subscribe to message get requests
    err = mr.eventBus.Subscribe("message.get", "message-runtime", func(ctx context.Context, data []byte) {
        var event events.MessageGetRequestEvent
        if err := json.Unmarshal(data, &event); err != nil {
            log.Printf("Failed to unmarshal message get event: %v", err)
            return
        }
        mr.handler.HandleMessageGet(context.Background(), event)
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

    // Initialize message store
    messageStore := store.NewMessageStore(redisURL)

    // Create message runtime
    runtime := NewMessageRuntime(eventBus, messageStore)

    // Start subscriptions
    if err := runtime.Start(); err != nil {
        log.Fatalf("Failed to start message runtime: %v", err)
    }

    log.Println("Message Runtime started successfully")

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down Message Runtime...")

    // Graceful shutdown
    _, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Close connections
    eventBus.Close()
    messageStore.Close()

    log.Println("Message Runtime stopped")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}