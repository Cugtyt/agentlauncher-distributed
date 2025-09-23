package store

import (
	"encoding/json"
	"fmt"
	"time"
)

type TaskData struct {
	AgentID string `json:"agent_id"`
	Task    string `json:"task"`
	Status  string `json:"status"` // "pending", "success", "failed"
	Result  string `json:"result,omitempty"`
}

type TaskStore struct {
	redis *RedisClient
}

func NewTaskStore(redisURL string) (*TaskStore, error) {
	redisClient, err := NewRedisClient(redisURL)
	if err != nil {
		return nil, err
	}
	return &TaskStore{
		redis: redisClient,
	}, nil
}

func (ts *TaskStore) taskKey(agentID string) string {
	return fmt.Sprintf("task:%s", agentID)
}

func (ts *TaskStore) CreateTaskPending(agentID, task string) error {
	taskData := TaskData{
		AgentID: agentID,
		Task:    task,
		Status:  "pending",
	}

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	if err := ts.redis.Set(ts.taskKey(agentID), string(jsonData), 12*time.Hour); err != nil {
		return fmt.Errorf("failed to create pending task: %w", err)
	}

	return nil
}

func (ts *TaskStore) CreateTaskSuccess(agentID, result string) error {
	exists, err := ts.TaskExists(agentID)
	if err != nil {
		return fmt.Errorf("failed to check task existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("task does not exist for agent %s", agentID)
	}

	existingTask, err := ts.GetTask(agentID)
	if err != nil {
		return fmt.Errorf("failed to get existing task: %w", err)
	}

	taskData := TaskData{
		AgentID: existingTask.AgentID,
		Task:    existingTask.Task,
		Status:  "success",
		Result:  result,
	}

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	if err := ts.redis.Set(ts.taskKey(agentID), string(jsonData), 12*time.Hour); err != nil {
		return fmt.Errorf("failed to create success task: %w", err)
	}

	return nil
}

func (ts *TaskStore) CreateTaskFailed(agentID, errorMsg string) error {
	exists, err := ts.TaskExists(agentID)
	if err != nil {
		return fmt.Errorf("failed to check task existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("task does not exist for agent %s", agentID)
	}

	existingTask, err := ts.GetTask(agentID)
	if err != nil {
		return fmt.Errorf("failed to get existing task: %w", err)
	}

	taskData := TaskData{
		AgentID: existingTask.AgentID,
		Task:    existingTask.Task,
		Status:  "failed",
		Result:  errorMsg,
	}

	jsonData, err := json.Marshal(taskData)
	if err != nil {
		return fmt.Errorf("failed to marshal task data: %w", err)
	}

	if err := ts.redis.Set(ts.taskKey(agentID), string(jsonData), 12*time.Hour); err != nil {
		return fmt.Errorf("failed to create failed task: %w", err)
	}

	return nil
}

func (ts *TaskStore) GetTask(agentID string) (*TaskData, error) {
	data, err := ts.redis.Get(ts.taskKey(agentID))
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	var task TaskData
	if err := json.Unmarshal([]byte(data), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task data: %w", err)
	}

	return &task, nil
}

func (ts *TaskStore) DeleteTask(agentID string) error {
	if err := ts.redis.Del(ts.taskKey(agentID)); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

func (ts *TaskStore) TaskExists(agentID string) (bool, error) {
	exists, err := ts.redis.Exists(ts.taskKey(agentID))
	if err != nil {
		return false, fmt.Errorf("failed to check if task exists: %w", err)
	}
	return exists > 0, nil
}

func (ts *TaskStore) HealthCheck() error {
	return ts.redis.Ping()
}

func (ts *TaskStore) Close() error {
	return ts.redis.Close()
}
