package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
	"github.com/cugtyt/agentlauncher-distributed/internal/events"
	"github.com/cugtyt/agentlauncher-distributed/internal/handlers"
	"github.com/cugtyt/agentlauncher-distributed/internal/llminterface"
	"github.com/cugtyt/agentlauncher-distributed/internal/runtimes"
)

type LLMRuntime struct {
	eventBus *eventbus.DistributedEventBus
	handler  *handlers.LLMHandler
}

func NewLLMRuntime() (*LLMRuntime, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		return nil, fmt.Errorf("NATS_URL environment variable is required")
	}

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	llmProcessor := func(messages []llminterface.Message, tools llminterface.RequestToolList, agentID string, eb eventbus.EventBus) ([]llminterface.Message, error) {
		log.Printf("[%s] Processing %d messages with %d tools", agentID, len(messages), len(tools))

		response := []llminterface.Message{
			llminterface.NewAssistantMessage("Hello from LLM processor"),
		}
		return response, nil
	}

	handler := handlers.NewLLMHandler(eventBus, llmProcessor)

	return &LLMRuntime{
		eventBus: eventBus,
		handler:  handler,
	}, nil
}

func (lr *LLMRuntime) Close() error {
	lr.eventBus.Close()
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
