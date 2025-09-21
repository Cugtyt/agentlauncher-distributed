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
	"github.com/cugtyt/agentlauncher-distributed/internal/tools"
)

type ToolRuntime struct {
	eventBus     *eventbus.DistributedEventBus
	toolRegistry *tools.Registry
	handler      *handlers.ToolHandler
}

func NewToolRuntime() (*ToolRuntime, error) {
	natsURL := utils.GetEnv("NATS_URL", "nats://localhost:4222")

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	toolRegistry := tools.NewRegistry()
	toolRegistry.RegisterTool(tools.NewSearchTool())
	toolRegistry.RegisterTool(tools.NewWeatherTool())
	toolRegistry.RegisterTool(tools.NewCreateAgentTool(eventBus))

	handler := handlers.NewToolHandler(eventBus, toolRegistry)

	return &ToolRuntime{
		eventBus:     eventBus,
		toolRegistry: toolRegistry,
		handler:      handler,
	}, nil
}

func (tr *ToolRuntime) Close() error {
	tr.eventBus.Close()
	return nil
}

func (tr *ToolRuntime) Start() error {
	return tr.eventBus.Subscribe(events.ToolExecuteEventName, runtimes.ToolRuntimeQueueName, func(ctx context.Context, data []byte) {
		if event, ok := utils.UnmarshalEvent[events.ToolsExecRequestEvent](data, events.ToolExecuteEventName); ok {
			tr.handler.HandleToolExecution(ctx, event)
		}
	})
}

func main() {
	toolRuntime, err := NewToolRuntime()
	if err != nil {
		log.Fatalf("Failed to initialize tool runtime: %v", err)
	}

	if err := toolRuntime.Start(); err != nil {
		log.Fatalf("Failed to start tool runtime: %v", err)
	}

	log.Println("Tool Runtime started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Tool Runtime...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	toolRuntime.Close()

	select {
	case <-ctx.Done():
		log.Println("Shutdown timeout exceeded")
	default:
		log.Println("Tool Runtime stopped")
	}
}
