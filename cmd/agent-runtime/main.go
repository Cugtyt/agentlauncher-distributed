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
	"github.com/cugtyt/agentlauncher-distributed/internal/runtimes"
	"github.com/cugtyt/agentlauncher-distributed/internal/store"
)

type AgentRuntime struct {
	eventBus     *eventbus.DistributedEventBus
	agentStore   *store.AgentStore
	messageStore *store.MessageStore
	handler      *handlers.AgentHandler
}

func NewAgentRuntime() (*AgentRuntime, error) {
	natsURL := utils.GetEnv("NATS_URL", "nats://localhost:4222")
	redisURL := utils.GetEnv("REDIS_URL", "redis://localhost:6379")

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	agentStore := store.NewAgentStore(redisURL)
	messageStore := store.NewMessageStore(redisURL)

	handler := handlers.NewAgentHandler(eventBus, agentStore, messageStore)

	return &AgentRuntime{
		eventBus:     eventBus,
		agentStore:   agentStore,
		messageStore: messageStore,
		handler:      handler,
	}, nil
}

func (ar *AgentRuntime) Close() error {
	ar.eventBus.Close()
	ar.agentStore.Close()
	ar.messageStore.Close()
	return nil
}

func (ar *AgentRuntime) Start() error {
	err := ar.eventBus.Subscribe(events.TaskCreateEventName, runtimes.AgentRuntimeQueueName, func(ctx context.Context, data []byte) {
		if event, ok := utils.UnmarshalEvent[events.TaskCreateEvent](data, events.TaskCreateEventName); ok {
			ar.handler.HandleTaskCreate(ctx, event)
		}
	})
	if err != nil {
		return err
	}

	err = ar.eventBus.Subscribe(events.AgentCreateEventName, runtimes.AgentRuntimeQueueName, func(ctx context.Context, data []byte) {
		if event, ok := utils.UnmarshalEvent[events.AgentCreateEvent](data, events.AgentCreateEventName); ok {
			ar.handler.HandleAgentCreate(ctx, event)
		}
	})
	if err != nil {
		return err
	}

	err = ar.eventBus.Subscribe(events.AgentStartEventName, runtimes.AgentRuntimeQueueName, func(ctx context.Context, data []byte) {
		if event, ok := utils.UnmarshalEvent[events.AgentStartEvent](data, events.AgentStartEventName); ok {
			ar.handler.HandleAgentStart(ctx, event)
		}
	})
	if err != nil {
		return err
	}

	err = ar.eventBus.Subscribe(events.LLMResponseEventName, runtimes.AgentRuntimeQueueName, func(ctx context.Context, data []byte) {
		if event, ok := utils.UnmarshalEvent[events.LLMResponseEvent](data, events.LLMResponseEventName); ok {
			ar.handler.HandleLLMResponse(ctx, event)
		}
	})
	if err != nil {
		return err
	}

	err = ar.eventBus.Subscribe(events.ToolResultEventName, runtimes.AgentRuntimeQueueName, func(ctx context.Context, data []byte) {
		if event, ok := utils.UnmarshalEvent[events.ToolsExecResultsEvent](data, events.ToolResultEventName); ok {
			ar.handler.HandleToolResult(ctx, event)
		}
	})

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
