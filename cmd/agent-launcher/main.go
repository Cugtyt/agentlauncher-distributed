package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/google/uuid"
    "github.com/gorilla/mux"
    "github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
    "github.com/cugtyt/agentlauncher-distributed/internal/events"
)

type AgentLauncher struct {
    eventBus *eventbus.DistributedEventBus
}

type CreateTaskRequest struct {
    Task        string            `json:"task"`
    Context     string            `json:"context,omitempty"`
    Tools       []string          `json:"tools,omitempty"`
    MaxSteps    int               `json:"max_steps,omitempty"`
    Metadata    map[string]string `json:"metadata,omitempty"`
}

type CreateTaskResponse struct {
    AgentID   string `json:"agent_id"`
    Status    string `json:"status"`
    Message   string `json:"message"`
}

func NewAgentLauncher(eventBus *eventbus.DistributedEventBus) *AgentLauncher {
    return &AgentLauncher{
        eventBus: eventBus,
    }
}

func (al *AgentLauncher) createTaskHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Generate unique agent ID
    agentID := uuid.New().String()

    // Create task event
    taskEvent := events.TaskCreateEvent{
        AgentID:  agentID,
        Task:     req.Task,
        Context:  req.Context,
        Tools:    req.Tools,
        MaxSteps: req.MaxSteps,
        Metadata: req.Metadata,
    }

    // Emit task creation event
    if err := al.eventBus.Emit("task.create", taskEvent); err != nil {
        log.Printf("Failed to emit task event: %v", err)
        http.Error(w, "Failed to create task", http.StatusInternalServerError)
        return
    }

    // Return response
    response := CreateTaskResponse{
        AgentID: agentID,
        Status:  "created",
        Message: "Task created successfully",
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func (al *AgentLauncher) healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}

func main() {
    natsURL := getEnv("NATS_URL", "nats://localhost:4222")
    port := getEnv("PORT", "8080")

    // Initialize event bus
    eventBus, err := eventbus.NewDistributedEventBus(natsURL)
    if err != nil {
        log.Fatalf("Failed to initialize event bus: %v", err)
    }

    // Create agent launcher
    launcher := NewAgentLauncher(eventBus)

    // Setup HTTP routes
    router := mux.NewRouter()
    router.HandleFunc("/tasks", launcher.createTaskHandler).Methods("POST")
    router.HandleFunc("/health", launcher.healthHandler).Methods("GET")

    // Setup HTTP server
    server := &http.Server{
        Addr:    ":" + port,
        Handler: router,
    }

    // Start server in a goroutine
    go func() {
        log.Printf("Agent Launcher starting on port %s", port)
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    log.Println("Shutting down Agent Launcher...")

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
    }

    eventBus.Close()
    log.Println("Agent Launcher stopped")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}