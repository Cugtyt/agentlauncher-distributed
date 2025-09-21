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
	"github.com/cugtyt/agentlauncher-distributed/internal/tools"
)

type ToolRuntime struct {
	eventBus     *eventbus.DistributedEventBus
	toolRegistry *tools.Registry
	handler      *handlers.ToolHandler
}

func NewToolRuntime(eventBus *eventbus.DistributedEventBus, toolRegistry *tools.Registry) *ToolRuntime {
	handler := handlers.NewToolHandler(eventBus, toolRegistry)

	return &ToolRuntime{
		eventBus:     eventBus,
		toolRegistry: toolRegistry,
		handler:      handler,
	}
}

func (tr *ToolRuntime) Start() error {
	// Subscribe to tool execution request events
	return tr.eventBus.Subscribe("tool.execute", "tool-runtime", func(ctx context.Context, data []byte) {
		var event events.ToolsExecRequestEvent
		if err := json.Unmarshal(data, &event); err != nil {
			log.Printf("Failed to unmarshal tool execute event: %v", err)
			return
		}
		tr.handler.HandleToolExecution(context.Background(), event)
	})
}

func main() {
	natsURL := getEnv("NATS_URL", "nats://localhost:4222")

	// Initialize event bus
	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		log.Fatalf("Failed to initialize event bus: %v", err)
	}

	// Initialize tool registry and register tools
	toolRegistry := tools.NewRegistry()

	// Register available tools
	toolRegistry.Register("search", tools.NewSearchTool())
	toolRegistry.Register("weather", tools.NewWeatherTool())
	toolRegistry.Register("create_agent", tools.NewCreateAgentTool(eventBus))

	// Create tool runtime
	runtime := NewToolRuntime(eventBus, toolRegistry)

	// Start subscriptions
	if err := runtime.Start(); err != nil {
		log.Fatalf("Failed to start tool runtime: %v", err)
	}

	log.Println("Tool Runtime started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Tool Runtime...")

	// Graceful shutdown
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Close connections
	eventBus.Close()

	log.Println("Tool Runtime stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
