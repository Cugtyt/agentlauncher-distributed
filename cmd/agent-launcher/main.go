package main

import (
	"bytes"
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
	"github.com/cugtyt/agentlauncher-distributed/internal/store"
	"github.com/cugtyt/agentlauncher-distributed/internal/utils"
)

const (
	StatusSuccess   = "success"
	StatusFailed    = "failed"
	StatusPending   = "pending"
	StatusCompleted = "completed"
)

type AgentLauncher struct {
	eventBus       *eventbus.DistributedEventBus
	handler        *handlers.LauncherHandler
	taskStore      *store.TaskStore
	toolRuntimeURL string
}

type CreateTaskRequest struct {
	Task         string                 `json:"task"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	Conversation []llminterface.Message `json:"conversation,omitempty"`
	Tools        []string               `json:"tools,omitempty"`
}

type CreateTaskResponse struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
}

type GetResultResponse struct {
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
	Result  string `json:"result,omitempty"`
	Message string `json:"message,omitempty"`
}

func NewAgentLauncher() (*AgentLauncher, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		return nil, fmt.Errorf("NATS_URL environment variable is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL environment variable is required")
	}

	toolRuntimeURL := os.Getenv("TOOL_RUNTIME_URL")
	if toolRuntimeURL == "" {
		return nil, fmt.Errorf("TOOL_RUNTIME_URL environment variable is required")
	}

	eventBus, err := eventbus.NewDistributedEventBus(natsURL)
	if err != nil {
		return nil, err
	}

	taskStore, err := store.NewTaskStore(redisURL)
	if err != nil {
		return nil, err
	}

	handler := handlers.NewLauncherHandler(taskStore)

	return &AgentLauncher{
		eventBus:       eventBus,
		handler:        handler,
		taskStore:      taskStore,
		toolRuntimeURL: toolRuntimeURL,
	}, nil
}

func (al *AgentLauncher) Start() error {
	if err := eventbus.Subscribe(al.eventBus, events.TaskFinishEventName, runtimes.AgentLauncherQueueName, al.handler.HandleTaskFinish); err != nil {
		return err
	}
	return eventbus.Subscribe(al.eventBus, events.TaskErrorEventName, runtimes.AgentLauncherQueueName, al.handler.HandleTaskError)
}

func (al *AgentLauncher) Close() error {
	return al.eventBus.Close()
}

func (al *AgentLauncher) getToolSchemas(toolNames []string) ([]llminterface.ToolSchema, error) {
	reqBody := struct {
		Tools []string `json:"tools"`
	}{
		Tools: toolNames,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(al.toolRuntimeURL+"/schemas", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to query tool runtime: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tool runtime returned status %d", resp.StatusCode)
	}

	var response struct {
		Schemas []llminterface.ToolSchema `json:"schemas"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return response.Schemas, nil
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

	agentID := utils.CreatePrimaryAgentID()

	if err := al.taskStore.CreateTaskPending(agentID, req.Task); err != nil {
		log.Printf("Failed to create task in store: %v", err)
		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	toolSchemas, err := al.getToolSchemas(req.Tools)
	if err != nil {
		log.Printf("Failed to get tool schemas: %v", err)
		al.taskStore.DeleteTask(agentID)
		http.Error(w, "Failed to get tool schemas", http.StatusInternalServerError)
		return
	}

	taskEvent := events.TaskCreateEvent{
		AgentID:      agentID,
		Task:         req.Task,
		SystemPrompt: req.SystemPrompt,
		ToolSchemas:  toolSchemas,
		Conversation: req.Conversation,
	}

	if err := al.eventBus.Emit(taskEvent); err != nil {
		log.Printf("Failed to emit task event: %v", err)

		al.taskStore.DeleteTask(agentID)

		http.Error(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	response := CreateTaskResponse{
		AgentID: agentID,
		Status:  StatusPending,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (al *AgentLauncher) getResultHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		http.Error(w, "agent_id parameter is required", http.StatusBadRequest)
		return
	}

	task, err := al.taskStore.GetTask(agentID)
	if err != nil {
		response := GetResultResponse{
			AgentID: agentID,
			Status:  StatusFailed,
			Message: "Task not found",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	var response GetResultResponse
	if task.Result != "" {
		response = GetResultResponse{
			AgentID: agentID,
			Status:  StatusCompleted,
			Result:  task.Result,
		}
	} else {
		response = GetResultResponse{
			AgentID: agentID,
			Status:  StatusPending,
			Message: "Task still in progress",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (al *AgentLauncher) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := http.Get(al.toolRuntimeURL + "/health")
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		http.Error(w, "Tool runtime not ready", http.StatusServiceUnavailable)
		return
	}
	resp.Body.Close()

	if err := al.taskStore.HealthCheck(); err != nil {
		http.Error(w, "Redis not ready", http.StatusServiceUnavailable)
		return
	}

	if !al.eventBus.IsConnected() {
		http.Error(w, "NATS not ready", http.StatusServiceUnavailable)
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

	launcher, err := NewAgentLauncher()
	if err != nil {
		log.Fatalf("Failed to initialize agent launcher: %v", err)
	}

	if err := launcher.Start(); err != nil {
		log.Fatalf("Failed to start agent launcher: %v", err)
	}

	http.HandleFunc("/tasks", launcher.createTaskHandler)
	http.HandleFunc("/results", launcher.getResultHandler)
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
