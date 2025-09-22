package main

import (
	"context"
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
	eventBus     *eventbus.DistributedEventBus
	messageStore *store.MessageStore
	handler      *handlers.LLMHandler
}

func NewLLMRuntime() (*LLMRuntime, error) {
	natsURL := utils.GetEnv("NATS_URL", "nats://localhost:4222")
	redisURL := utils.GetEnv("REDIS_URL", "redis://localhost:6379")

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	messageStore := store.NewMessageStore(redisURL)

	llmProcessor := func(messages llminterface.RequestMessageList, tools llminterface.RequestToolList, agentID string, eb eventbus.EventBus) (llminterface.ResponseMessageList, error) {
		log.Printf("[%s] Processing %d messages with %d tools", agentID, len(messages), len(tools))

		response := llminterface.ResponseMessageList{
			llminterface.AssistantMessage{Content: "Hello from LLM processor"},
		}
		return response, nil
	}

	handler := handlers.NewLLMHandler(eventBus, llmProcessor)

	return &LLMRuntime{
		eventBus:     eventBus,
		messageStore: messageStore,
		handler:      handler,
	}, nil
}

func (lr *LLMRuntime) Close() error {
	lr.eventBus.Close()
	lr.messageStore.Close()
	return nil
}

func (lr *LLMRuntime) Start() error {
	if err := eventbus.Subscribe(lr.eventBus, events.LLMRequestEventName, runtimes.LLMRuntimeQueueName, lr.handler.HandleLLMRequest); err != nil {
		return err
	}

	return eventbus.Subscribe(lr.eventBus, events.LLMErrorEventName, runtimes.LLMRuntimeQueueName, lr.handler.HandleLLMRuntimeError)
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
