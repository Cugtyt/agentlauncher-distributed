package handlers

import (
	"context"
	"log"

	"github.com/cugtyt/agentlauncher-distributed/internal/events"
	"github.com/cugtyt/agentlauncher-distributed/internal/store"
)

type LauncherHandler struct {
	taskStore *store.TaskStore
}

func NewLauncherHandler(taskStore *store.TaskStore) *LauncherHandler {
	return &LauncherHandler{
		taskStore: taskStore,
	}
}

func (h *LauncherHandler) HandleTaskFinish(ctx context.Context, event events.TaskFinishEvent) {
	if err := h.taskStore.CreateTaskSuccess(event.AgentID, event.Result); err != nil {
		log.Printf("Failed to update task success for agent %s: %v", event.AgentID, err)
	}
}

func (h *LauncherHandler) HandleTaskError(ctx context.Context, event events.TaskErrorEvent) {
	if err := h.taskStore.CreateTaskFailed(event.AgentID, event.Error); err != nil {
		log.Printf("Failed to update task failure for agent %s: %v", event.AgentID, err)
	}
}
