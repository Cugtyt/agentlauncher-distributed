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

type MessageRuntime struct {
	eventBus     eventbus.EventBus
	messageStore *store.MessageStore
	handler      *handlers.MessageHandler
}

func NewMessageRuntime() (*MessageRuntime, error) {
	natsURL := utils.GetEnv("NATS_URL", "nats://localhost:4222")
	redisURL := utils.GetEnv("REDIS_URL", "redis://localhost:6379")

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	messageStore := store.NewMessageStore(redisURL)
	handler := handlers.NewMessageHandler(eventBus, messageStore)

	return &MessageRuntime{
		eventBus:     eventBus,
		messageStore: messageStore,
		handler:      handler,
	}, nil
}

func (mr *MessageRuntime) Close() error {
	mr.eventBus.Close()
	mr.messageStore.Close()
	return nil
}

func (mr *MessageRuntime) Start() error {
	err := mr.eventBus.Subscribe(events.MessageAddEventName, runtimes.MessageRuntimeQueueName, func(ctx context.Context, data []byte) {
		if event, ok := utils.UnmarshalEvent[events.MessagesAddEvent](data, events.MessageAddEventName); ok {
			mr.handler.HandleMessageAdd(ctx, event)
		}
	})
	if err != nil {
		return err
	}

	err = mr.eventBus.Subscribe(events.MessageGetEventName, runtimes.MessageRuntimeQueueName, func(ctx context.Context, data []byte) {
		if event, ok := utils.UnmarshalEvent[events.MessageGetRequestEvent](data, events.MessageGetEventName); ok {
			mr.handler.HandleMessageGet(ctx, event)
		}
	})

	return err
}

func main() {
	messageRuntime, err := NewMessageRuntime()
	if err != nil {
		log.Fatalf("Failed to initialize message runtime: %v", err)
	}

	if err := messageRuntime.Start(); err != nil {
		log.Fatalf("Failed to start message runtime: %v", err)
	}

	log.Println("Message Runtime started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Message Runtime...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messageRuntime.Close()

	select {
	case <-ctx.Done():
		log.Println("Shutdown timeout exceeded")
	default:
		log.Println("Message Runtime stopped")
	}
}
