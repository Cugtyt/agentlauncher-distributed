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
	"github.com/cugtyt/agentlauncher-distributed/internal/runtimes"
	"github.com/cugtyt/agentlauncher-distributed/internal/store"
)

type AgentRuntime struct {
	eventBus   *eventbus.DistributedEventBus
	agentStore *store.AgentStore
	handler    *handlers.AgentHandler
}

func NewAgentRuntime() (*AgentRuntime, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		return nil, fmt.Errorf("NATS_URL environment variable is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable is required")
	}

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	agentStore, err := store.NewAgentStore(redisURL)
	if err != nil {
		return nil, err
	}

	handler := handlers.NewAgentHandler(eventBus, agentStore)

	return &AgentRuntime{
		eventBus:   eventBus,
		agentStore: agentStore,
		handler:    handler,
	}, nil
}

func (ar *AgentRuntime) Close() error {
	ar.eventBus.Close()
	ar.agentStore.Close()
	return nil
}

func (ar *AgentRuntime) Start() error {
	err := eventbus.Subscribe(ar.eventBus, events.TaskCreateEventName, runtimes.AgentRuntimeQueueName, ar.handler.HandleTaskCreate)
	if err != nil {
		return err
	}

	err = eventbus.Subscribe(ar.eventBus, events.AgentCreateEventName, runtimes.AgentRuntimeQueueName, ar.handler.HandleAgentCreate)
	if err != nil {
		return err
	}

	err = eventbus.Subscribe(ar.eventBus, events.AgentStartEventName, runtimes.AgentRuntimeQueueName, ar.handler.HandleAgentStart)
	if err != nil {
		return err
	}

	err = eventbus.Subscribe(ar.eventBus, events.LLMResponseEventName, runtimes.AgentRuntimeQueueName, ar.handler.HandleLLMResponse)
	if err != nil {
		return err
	}

	err = eventbus.Subscribe(ar.eventBus, events.ToolExecResultsEventName, runtimes.AgentRuntimeQueueName, ar.handler.HandleToolResult)
	if err != nil {
		return err
	}

	err = eventbus.Subscribe(ar.eventBus, events.AgentFinishEventName, runtimes.AgentRuntimeQueueName, ar.handler.HandleAgentFinish)
	if err != nil {
		return err
	}

	err = eventbus.Subscribe(ar.eventBus, events.AgentErrorEventName, runtimes.AgentRuntimeQueueName, ar.handler.HandleAgentError)
	if err != nil {
		return err
	}

	err = eventbus.Subscribe(ar.eventBus, events.AgentDeletedEventName, runtimes.AgentRuntimeQueueName, ar.handler.HandleAgentDeleted)

	return err
}

func main() {
	runtime, err := NewAgentRuntime()
	if err != nil {
		log.Fatalf("Failed to initialize agent runtime: %v", err)
	}

	if err := runtime.Start(); err != nil {
		log.Fatalf("Failed to start agent runtime: %v", err)
	}

	log.Println("Agent Runtime started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Agent Runtime...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	runtime.Close()

	select {
	case <-ctx.Done():
		log.Println("Shutdown timeout exceeded")
	default:
		log.Println("Agent Runtime stopped")
	}
}
