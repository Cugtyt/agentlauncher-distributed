package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cugtyt/agentlauncher-distributed/cmd/utils"
	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
	"github.com/cugtyt/agentlauncher-distributed/internal/events"
	"github.com/cugtyt/agentlauncher-distributed/internal/handlers"
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
	"github.com/cugtyt/agentlauncher-distributed/internal/runtimes"
	"github.com/cugtyt/agentlauncher-distributed/internal/store"
)

type LLMRuntime struct {
	eventBus     eventbus.EventBus
	messageStore *store.MessageStore
	llmClient    *llminterface.OpenAIClient
	handler      *handlers.LLMHandler
}

func NewLLMRuntime() (*LLMRuntime, error) {
	natsURL := utils.GetEnv("NATS_URL", "nats://localhost:4222")
	redisURL := utils.GetEnv("REDIS_URL", "redis://localhost:6379")
	openaiAPIKey := utils.GetEnv("OPENAI_API_KEY", "")

	if openaiAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	messageStore := store.NewMessageStore(redisURL)
	llmClient := llminterface.NewOpenAIClient(openaiAPIKey)
	handler := handlers.NewLLMHandler(eventBus, messageStore, llmClient)

	return &LLMRuntime{
		eventBus:     eventBus,
		messageStore: messageStore,
		llmClient:    llmClient,
		handler:      handler,
	}, nil
}

func (lr *LLMRuntime) Close() error {
	lr.eventBus.Close()
	lr.messageStore.Close()
	return nil
}

func (lr *LLMRuntime) Start() error {
	return lr.eventBus.Subscribe(events.LLMRequestEventName, runtimes.LLMRuntimeQueueName, func(ctx context.Context, data []byte) {
		if event, ok := utils.UnmarshalEvent[events.LLMRequestEvent](data, events.LLMRequestEventName); ok {
			lr.handler.HandleLLMRequest(ctx, event)
		}
	})
}

func main() {
	runtime, err := NewLLMRuntime()
	if err != nil {
		log.Fatalf("Failed to initialize LLM runtime: %v", err)
	}

	if err := runtime.Start(); err != nil {
		log.Fatalf("Failed to start LLM runtime: %v", err)
	}

	log.Println("LLM Runtime started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down LLM Runtime...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runtime.Close()

	select {
	case <-ctx.Done():
		log.Println("Shutdown timeout exceeded")
	default:
		log.Println("LLM Runtime stopped")
	}
}
