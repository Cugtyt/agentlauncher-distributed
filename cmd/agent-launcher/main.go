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

	"github.com/cugtyt/agentlauncher-distributed/cmd/utils"
	"github.com/cugtyt/agentlauncher-distributed/internal/eventbus"
	"github.com/cugtyt/agentlauncher-distributed/internal/events"
	"github.com/google/uuid"
)

const (
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

type AgentLauncher struct {
	eventBus eventbus.EventBus
}

type CreateTaskRequest struct {
	Task         string   `json:"task"`
	Context      string   `json:"context,omitempty"`
	ToolNameList []string `json:"tool_name_list,omitempty"`
}

type CreateTaskResponse struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func NewAgentLauncher() (*AgentLauncher, error) {
	natsURL := utils.GetEnv("NATS_URL", "nats://localhost:4222")

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	return &AgentLauncher{
		eventBus: eventBus,
	}, nil
}

func (al *AgentLauncher) Close() error {
	return al.eventBus.Close()
}

func (al *AgentLauncher) createTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	agentID := uuid.New().String()

	taskEvent := events.TaskCreateEvent{
		AgentID:   agentID,
		Task:      req.Task,
		Context:   req.Context,
		Timestamp: time.Now(),
	}

	if err := al.eventBus.Emit("task.create", taskEvent); err != nil {
		log.Printf("Failed to emit task event: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	response := CreateTaskResponse{
		AgentID: agentID,
		Status:  StatusSuccess,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (al *AgentLauncher) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	port := utils.GetEnv("PORT", "8080")

	launcher, err := NewAgentLauncher()
	if err != nil {
		log.Fatalf("Failed to initialize agent launcher: %v", err)
	}

	http.HandleFunc("/tasks", launcher.createTaskHandler)
	http.HandleFunc("/health", launcher.healthHandler)

	server := &http.Server{
		Addr: ":" + port,
	}

	go func() {
		log.Printf("Agent Launcher starting on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Agent Launcher...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	launcher.Close()
	log.Println("Agent Launcher stopped")
}
