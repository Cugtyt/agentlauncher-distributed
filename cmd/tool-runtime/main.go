package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

type ToolRuntime struct {
	eventBus *eventbus.DistributedEventBus
	handler  *handlers.ToolHandler
}

func NewToolRuntime() (*ToolRuntime, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		return nil, fmt.Errorf("NATS_URL environment variable is required")
	}

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	handler := handlers.NewToolHandler(eventBus)

	calculatorTool := handlers.Tool{
		ToolSchema: llminterface.ToolSchema{
			Name:        "calculator",
			Description: "Perform basic arithmetic operations",
			Parameters: []llminterface.ToolParamSchema{
				{
					Type:        "string",
					Name:        "operation",
					Description: "add, subtract, multiply, divide",
					Required:    true,
				},
				{
					Type:        "number",
					Name:        "a",
					Description: "First number",
					Required:    true,
				},
				{
					Type:        "number",
					Name:        "b",
					Description: "Second number",
					Required:    true,
				},
			},
		},
		Function: func(ctx context.Context, params map[string]interface{}) (string, error) {
			operation := params["operation"].(string)
			a := params["a"].(float64)
			b := params["b"].(float64)

			switch operation {
			case "add":
				return fmt.Sprintf("%.2f", a+b), nil
			case "subtract":
				return fmt.Sprintf("%.2f", a-b), nil
			case "multiply":
				return fmt.Sprintf("%.2f", a*b), nil
			case "divide":
				if b == 0 {
					return "", fmt.Errorf("division by zero")
				}
				return fmt.Sprintf("%.2f", a/b), nil
			default:
				return "", fmt.Errorf("unknown operation: %s", operation)
			}
		},
	}

	weatherTool := handlers.Tool{
		ToolSchema: llminterface.ToolSchema{
			Name:        "weather",
			Description: "Get weather information for a city",
			Parameters: []llminterface.ToolParamSchema{
				{
					Type:        "string",
					Name:        "city",
					Description: "City name",
					Required:    true,
				},
			},
		},
		Function: func(ctx context.Context, params map[string]interface{}) (string, error) {
			city := params["city"].(string)
			return fmt.Sprintf("Weather in %s: Sunny, 25°C", city), nil
		},
	}

	timeTool := handlers.Tool{
		ToolSchema: llminterface.ToolSchema{
			Name:        "current_time",
			Description: "Get current time",
			Parameters:  []llminterface.ToolParamSchema{},
		},
		Function: func(ctx context.Context, params map[string]interface{}) (string, error) {
			return time.Now().Format("2006-01-02 15:04:05"), nil
		},
	}

	randomTool := handlers.Tool{
		ToolSchema: llminterface.ToolSchema{
			Name:        "random_number",
			Description: "Generate a random number between min and max",
			Parameters: []llminterface.ToolParamSchema{
				{
					Type:        "number",
					Name:        "min",
					Description: "Minimum value",
					Required:    true,
				},
				{
					Type:        "number",
					Name:        "max",
					Description: "Maximum value",
					Required:    true,
				},
			},
		},
		Function: func(ctx context.Context, params map[string]interface{}) (string, error) {
			min := int(params["min"].(float64))
			max := int(params["max"].(float64))
			result := min + (time.Now().Nanosecond() % (max - min + 1))
			return fmt.Sprintf("%d", result), nil
		},
	}

	handler.Register(calculatorTool)
	handler.Register(weatherTool)
	handler.Register(timeTool)
	handler.Register(randomTool)

	return &ToolRuntime{
		eventBus: eventBus,
		handler:  handler,
	}, nil
}

func (tr *ToolRuntime) Close() error {
	tr.eventBus.Close()
	return nil
}

func (tr *ToolRuntime) Start() error {
	return eventbus.Subscribe(tr.eventBus, events.ToolExecRequestEventName, runtimes.ToolRuntimeQueueName, tr.handler.HandleToolExecution)
}

func (tr *ToolRuntime) getSchemasHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Tools []string `json:"tools"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var schemas []llminterface.ToolSchema
	if req.Tools == nil {
		schemas = tr.handler.GetAllToolSchemas()
	} else {
		for _, toolName := range req.Tools {
			if tool, err := tr.handler.GetTool(toolName); err == nil {
				schemas = append(schemas, tool.ToolSchema)
			}
		}
	}

	response := struct {
		Schemas []llminterface.ToolSchema `json:"schemas"`
	}{
		Schemas: schemas,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (tr *ToolRuntime) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is required")
	}

	toolRuntime, err := NewToolRuntime()
	if err != nil {
		log.Fatalf("Failed to initialize tool runtime: %v", err)
	}

	if err := toolRuntime.Start(); err != nil {
		log.Fatalf("Failed to start tool runtime: %v", err)
	}

	http.HandleFunc("/schemas", toolRuntime.getSchemasHandler)
	http.HandleFunc("/health", toolRuntime.healthHandler)

	server := &http.Server{
		Addr: ":" + port,
	}

	go func() {
		log.Printf("Tool Runtime starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Println("Tool Runtime started successfully")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Tool Runtime...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	toolRuntime.Close()

	select {
	case <-ctx.Done():
		log.Println("Shutdown timeout exceeded")
	default:
		log.Println("Tool Runtime stopped")
	}
}
