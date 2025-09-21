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
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
	"github.com/cugtyt/agentlauncher-distributed/internal/store"
)

type LLMRuntime struct {
	eventBus     *eventbus.DistributedEventBus
	messageStore *store.MessageStore
	llmClient    *llminterface.OpenAIClient
	handler      *handlers.LLMHandler
}

func NewLLMRuntime(eventBus *eventbus.DistributedEventBus, messageStore *store.MessageStore, llmClient *llminterface.OpenAIClient) *LLMRuntime {
	handler := handlers.NewLLMHandler(eventBus, messageStore, llmClient)

	return &LLMRuntime{
		eventBus:     eventBus,
		messageStore: messageStore,
		llmClient:    llmClient,
		handler:      handler,
	}
}

func (lr *LLMRuntime) Start() error {
	// Subscribe to LLM request events
	return lr.eventBus.Subscribe("llm.request", "llm-runtime", func(ctx context.Context, data []byte) {
		var event events.LLMRequestEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("Failed to unmarshal LLM request event: %v", err)
			return
		}
		lr.handler.HandleLLMRequest(context.Background(), event)
	})
}

func main() {
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	openaiAPIKey := getEnv("OPENAI_API_KEY", "")

	if openaiAPIKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Initialize event bus
	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		log.Fatalf("Failed to initialize event bus: %v", err)
	}

	// Initialize message store
	messageStore := store.NewMessageStore(redisURL)

	// Initialize LLM client
	llmClient := llminterface.NewOpenAIClient(openaiAPIKey)

	// Create LLM runtime
	runtime := NewLLMRuntime(eventBus, messageStore, llmClient)

	// Start subscriptions
	if err := runtime.Start(); err != nil {
		log.Fatalf("Failed to start LLM runtime: %v", err)
	}

	log.Println("LLM Runtime started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down LLM Runtime...")

	// Graceful shutdown
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close connections
	eventBus.Close()
	messageStore.Close()

	log.Println("LLM Runtime stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
